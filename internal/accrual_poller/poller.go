package accrual_poller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/entities"
	"go.uber.org/zap"
)

type IOrderService interface {
	GetOrdersToPoll(ctx context.Context) ([]entities.Order, error)
}

type ITransactionService interface {
	CreateAccrual(ctx context.Context, order entities.Order) error
}

type accrualResult struct {
	updated    *entities.Order
	retryAfter time.Duration
	err        error
}

type AccrualPoller struct {
	client             *resty.Client
	cfg                *config.Config
	logger             *zap.Logger
	orderService       IOrderService
	transactionService ITransactionService
	workerCount        int
}

func NewAccrualPoller(
	cfg *config.Config,
	logger *zap.Logger,
	orderService IOrderService,
	transactionService ITransactionService,
	workerCount int,
) *AccrualPoller {
	client := resty.New().SetBaseURL(cfg.AccrualSystemAddr)
	return &AccrualPoller{
		client:             client,
		cfg:                cfg,
		logger:             logger,
		orderService:       orderService,
		transactionService: transactionService,
		workerCount:        workerCount,
	}
}

func (p *AccrualPoller) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.cfg.AccrualPollInterval) * time.Second)
	defer ticker.Stop()

	p.logger.Info("started AccrualPoller", zap.String("address", p.client.BaseURL), zap.Int("workers", p.workerCount))
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

func (p *AccrualPoller) accrualWorker(ctx context.Context, id int, jobsCh <-chan entities.Order, resultsCh chan<- accrualResult) {
	defer p.logger.Info("accrual worker done", zap.Int("id", id))
	p.logger.Info("accrual worker started", zap.Int("id", id))

	for order := range jobsCh {
		select {
		case <-ctx.Done():
			return
		default:
		}

		resp, err := GetAccrualOrder(p.client, order.Number)
		if err != nil {
			var rateLimitErr *RetryAfterError
			if errors.As(err, &rateLimitErr) {
				resultsCh <- accrualResult{retryAfter: rateLimitErr.RetryAfter}
				return
			}
			if errors.Is(err, ErrNotRegistered) {
				continue
			}
			resultsCh <- accrualResult{err: fmt.Errorf("order %s: %w", order.Number, err)}
			continue
		}

		resultsCh <- accrualResult{updated: mapStatus(order, resp)}
	}
}

func (p *AccrualPoller) poll(ctx context.Context) error {
	orders, err := p.orderService.GetOrdersToPoll(ctx)
	if err != nil {
		return fmt.Errorf("GetOrdersToPoll: %w", err)
	}
	if len(orders) == 0 {
		return nil
	}

	pollCtx, cancelPoll := context.WithCancel(ctx)
	defer cancelPoll()

	jobsCh := make(chan entities.Order, len(orders))
	resultsCh := make(chan accrualResult, len(orders))

	var wg sync.WaitGroup
	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p.accrualWorker(pollCtx, id, jobsCh, resultsCh)
		}(i)
	}

	go func() {
		defer close(jobsCh)
		for _, order := range orders {
			select {
			case <-pollCtx.Done():
				return
			case jobsCh <- order:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	for result := range resultsCh {
		if result.retryAfter > 0 {
			p.logger.Warn("accrual rate limit, pausing", zap.Duration("retry_after", result.retryAfter))
			cancelPoll()
			select {
			case <-ctx.Done():
			case <-time.After(result.retryAfter):
			}
			for range resultsCh {
			}
			return nil
		}
		if result.err != nil {
			p.logger.Error("accrual error", zap.Error(result.err))
			continue
		}
		if result.updated == nil {
			continue
		}
		if err := p.transactionService.CreateAccrual(ctx, *result.updated); err != nil {
			p.logger.Error("applyUpdate failed", zap.String("order", result.updated.Number), zap.Error(err))
		}
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
