package accrual_poller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/repository"
	"go.uber.org/zap"
)

const pollInterval = 2 * time.Second

type AccrualPoller struct {
	client *resty.Client
	uow    repository.UnitOfWork
	logger *zap.Logger
}

func NewAccrualPoller(address string, uow repository.UnitOfWork, logger *zap.Logger) *AccrualPoller {
	client := resty.New().SetBaseURL(fmt.Sprintf("http://%s", address))
	return &AccrualPoller{client: client, uow: uow, logger: logger}
}

func (p *AccrualPoller) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	p.logger.Info("Started AccrualPoller", zap.String("address", p.client.BaseURL))
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.poll(ctx); err != nil {
				p.logger.Error("poller error", zap.Error(err))
			}
		}
	}
}

func (p *AccrualPoller) poll(ctx context.Context) error {
	orders, err := p.uow.Orders().GetOrdersToPoll(ctx)
	if err != nil {
		return fmt.Errorf("GetOrdersToPoll: %w", err)
	}
	if len(orders) == 0 {
		return nil
	}

	var toUpdate []entities.Order
	for _, order := range orders {
		resp, err := GetAccrualOrder(p.client, order.Number)
		if err != nil {
			var rateLimitErr *RetryAfterError
			if errors.As(err, &rateLimitErr) {
				p.logger.Warn("accrual rate limit, pausing", zap.Duration("retry_after", rateLimitErr.RetryAfter))
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(rateLimitErr.RetryAfter):
				}
				break
			}
			if errors.Is(err, ErrNotRegistered) {
				continue
			}
			p.logger.Error("accrual error", zap.String("order", order.Number), zap.Error(err))
			continue
		}

		updated := mapStatus(order, resp)
		if updated == nil {
			continue
		}
		toUpdate = append(toUpdate, *updated)
	}

	if len(toUpdate) == 0 {
		return nil
	}
	if err := p.uow.Orders().UpdateOrders(ctx, toUpdate); err != nil {
		return fmt.Errorf("UpdateOrders: %w", err)
	}
	return nil
}

func mapStatus(order entities.Order, resp *AccrualResponse) *entities.Order {
	var newStatus entities.OrderStatus
	switch resp.Status {
	case "REGISTERED":
		newStatus = entities.StatusNew
	case "PROCESSING":
		newStatus = entities.StatusProcessing
	case "INVALID":
		newStatus = entities.StatusInvalid
	case "PROCESSED":
		newStatus = entities.StatusProcessed
	default:
		return nil
	}

	if order.Status == newStatus && order.Accrual == resp.Accrual {
		return nil
	}

	order.Status = newStatus
	order.Accrual = resp.Accrual
	return &order
}
