package accrualpoller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/entities"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// IOrderService интерфейс для получения заказов, ожидающих обработки в accrual системе.
type IOrderService interface {
	GetOrdersToPoll(ctx context.Context) ([]entities.Order, error)
}

// ITransactionService интерфейс для сохранения результата начисления баллов по заказу.
type ITransactionService interface {
	CreateAccrual(ctx context.Context, order entities.Order) error
}

// accrualResult внутренняя структура для передачи результата от worker'а в основной цикл poll.
// Поля взаимоисключающи: задан ровно один из updated, retryAfter или err.
type accrualResult struct {
	updated    *entities.Order
	retryAfter time.Duration
	err        error
}

// AccrualPoller периодически опрашивает accrual систему для обновления статусов заказов.
// Обработка заказов выполняется параллельно через пул worker'ов.
type AccrualPoller struct {
	client             *resty.Client
	cfg                *config.Config
	logger             *zap.Logger
	orderService       IOrderService
	transactionService ITransactionService
	workerCount        int
}

// NewAccrualPoller создаёт новый AccrualPoller и инициализирует HTTP-клиент с базовым URL accrual системы.
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

// Run запускает цикл опроса accrual системы с интервалом из конфигурации.
// Блокирует выполнение до отмены ctx.
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

// accrualWorker читает заказы из jobsCh, запрашивает их статус в accrual системе
// и отправляет результат в resultsCh. При получении rate limit останавливается и завершает работу.
func (p *AccrualPoller) accrualWorker(ctx context.Context, id int, jobsCh <-chan entities.Order, resultsCh chan<- accrualResult) error {
	defer p.logger.Info("accrual worker done", zap.Int("id", id))
	p.logger.Info("accrual worker started", zap.Int("id", id))

	for order := range jobsCh {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		resp, err := GetAccrualOrder(p.client, order.Number)
		if err != nil {
			var rateLimitErr *RetryAfterError
			if errors.As(err, &rateLimitErr) {
				resultsCh <- accrualResult{retryAfter: rateLimitErr.RetryAfter}
				return nil
			}
			if errors.Is(err, ErrNotRegistered) {
				continue
			}
			resultsCh <- accrualResult{err: fmt.Errorf("order %s: %w", order.Number, err)}
			continue
		}

		resultsCh <- accrualResult{updated: mapStatus(order, resp)}
	}
	return nil
}

// poll выполняет один цикл опроса: загружает заказы, распределяет их по worker'ам
// и применяет полученные обновления. При rate limit от accrual системы делает паузу
// на указанное в заголовке Retry-After время.
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

	eg, egCtx := errgroup.WithContext(pollCtx)

	jobsCh := make(chan entities.Order, len(orders))
	resultsCh := make(chan accrualResult, len(orders))

	for i := 0; i < p.workerCount; i++ {
		id := i
		eg.Go(func() error {
			return p.accrualWorker(egCtx, id, jobsCh, resultsCh)
		})
	}

	eg.Go(func() error {
		defer close(jobsCh)
		for _, order := range orders {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			case jobsCh <- order:
			}
		}
		return nil
	})

	go func() {
		eg.Wait()
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

// mapStatus преобразует статус из ответа accrual системы в entities.OrderStatus
// и возвращает обновлённый Order. Если статус и сумма начисления не изменились — возвращает nil.
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
