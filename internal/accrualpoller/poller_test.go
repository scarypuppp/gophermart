package accrualpoller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// --- mocks ---

type mockOrderService struct {
	orders []entities.Order
	err    error
}

func (m *mockOrderService) GetOrdersToPoll(_ context.Context) ([]entities.Order, error) {
	return m.orders, m.err
}

type mockTransactionService struct {
	calls  atomic.Int32
	err    error
	lastFn func(order entities.Order)
}

func (m *mockTransactionService) CreateAccrual(_ context.Context, order entities.Order) error {
	m.calls.Add(1)
	if m.lastFn != nil {
		m.lastFn(order)
	}
	return m.err
}

// --- helpers ---

func newTestPoller(t *testing.T, srv *httptest.Server, os IOrderService, ts ITransactionService) *AccrualPoller {
	t.Helper()
	cfg := &config.Config{AccrualSystemAddr: srv.Listener.Addr().String(), AccrualPollInterval: 1}
	p := NewAccrualPoller(cfg, zap.NewNop(), os, ts, 2)
	// направляем клиент на тестовый сервер
	p.client = resty.New().SetBaseURL(srv.URL)
	return p
}

func accrualServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// --- GetAccrualOrder ---

func TestGetAccrualOrder_OK(t *testing.T) {
	accrual := 99.5
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AccrualResponse{Order: "123", Status: "PROCESSED", Accrual: &accrual})
	})
	client := resty.New().SetBaseURL(srv.URL)

	resp, err := GetAccrualOrder(client, "123")
	require.NoError(t, err)
	assert.Equal(t, "PROCESSED", resp.Status)
	assert.Equal(t, accrual, *resp.Accrual)
}

func TestGetAccrualOrder_NotRegistered(t *testing.T) {
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	client := resty.New().SetBaseURL(srv.URL)

	_, err := GetAccrualOrder(client, "123")
	assert.ErrorIs(t, err, ErrNotRegistered)
}

func TestGetAccrualOrder_RateLimit(t *testing.T) {
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	client := resty.New().SetBaseURL(srv.URL)

	_, err := GetAccrualOrder(client, "123")
	var rateLimitErr *RetryAfterError
	require.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, 2*time.Second, rateLimitErr.RetryAfter)
}

func TestGetAccrualOrder_UnexpectedStatus(t *testing.T) {
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	client := resty.New().SetBaseURL(srv.URL)

	_, err := GetAccrualOrder(client, "123")
	assert.Error(t, err)
}

// --- mapStatus ---

func TestMapStatus(t *testing.T) {
	order := entities.Order{Number: "1", Status: entities.StatusNew}

	t.Run("PROCESSED sets status and accrual", func(t *testing.T) {
		accrual := 10.0
		result := mapStatus(order, &AccrualResponse{Status: "PROCESSED", Accrual: &accrual})
		require.NotNil(t, result)
		assert.Equal(t, entities.StatusProcessed, result.Status)
		assert.Equal(t, &accrual, result.Accrual)
	})

	t.Run("unknown status returns nil", func(t *testing.T) {
		assert.Nil(t, mapStatus(order, &AccrualResponse{Status: "UNKNOWN"}))
	})

	t.Run("no change returns nil", func(t *testing.T) {
		same := entities.Order{Status: entities.StatusNew}
		assert.Nil(t, mapStatus(same, &AccrualResponse{Status: "REGISTERED"}))
	})

	t.Run("PROCESSING", func(t *testing.T) {
		result := mapStatus(order, &AccrualResponse{Status: "PROCESSING"})
		require.NotNil(t, result)
		assert.Equal(t, entities.StatusProcessing, result.Status)
	})

	t.Run("INVALID", func(t *testing.T) {
		result := mapStatus(order, &AccrualResponse{Status: "INVALID"})
		require.NotNil(t, result)
		assert.Equal(t, entities.StatusInvalid, result.Status)
	})
}

// --- poll ---

func TestPoll_EmptyOrders(t *testing.T) {
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {})
	p := newTestPoller(t, srv, &mockOrderService{}, &mockTransactionService{})

	err := p.poll(context.Background())
	assert.NoError(t, err)
}

func TestPoll_OrderServiceError(t *testing.T) {
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {})
	os := &mockOrderService{err: errors.New("db down")}
	p := newTestPoller(t, srv, os, &mockTransactionService{})

	err := p.poll(context.Background())
	assert.ErrorContains(t, err, "GetOrdersToPoll")
}

func TestPoll_ProcessedOrder_CallsCreateAccrual(t *testing.T) {
	accrual := 50.0
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AccrualResponse{Status: "PROCESSED", Accrual: &accrual})
	})

	os := &mockOrderService{orders: []entities.Order{{Number: "12345678903", Status: entities.StatusNew}}}
	ts := &mockTransactionService{}
	p := newTestPoller(t, srv, os, ts)

	err := p.poll(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int32(1), ts.calls.Load())
}

func TestPoll_NotRegistered_Skipped(t *testing.T) {
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	os := &mockOrderService{orders: []entities.Order{{Number: "12345678903", Status: entities.StatusNew}}}
	ts := &mockTransactionService{}
	p := newTestPoller(t, srv, os, ts)

	err := p.poll(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int32(0), ts.calls.Load())
}

func TestPoll_NoStatusChange_SkipsCreateAccrual(t *testing.T) {
	// REGISTERED маппится в StatusNew — статус не изменился, mapStatus вернёт nil
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AccrualResponse{Status: "REGISTERED"})
	})

	os := &mockOrderService{orders: []entities.Order{{Number: "12345678903", Status: entities.StatusNew}}}
	ts := &mockTransactionService{}
	p := newTestPoller(t, srv, os, ts)

	err := p.poll(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int32(0), ts.calls.Load())
}

func TestPoll_RateLimit_StopsEarly(t *testing.T) {
	var callCount atomic.Int32
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	})

	orders := make([]entities.Order, 5)
	for i := range orders {
		orders[i] = entities.Order{Number: fmt.Sprintf("order-%d", i), Status: entities.StatusNew}
	}
	os := &mockOrderService{orders: orders}
	ts := &mockTransactionService{}
	p := newTestPoller(t, srv, os, ts)

	// отменяем контекст, чтобы разблокировать ожидание Retry-After
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := p.poll(ctx)
	require.NoError(t, err)
	assert.Equal(t, int32(0), ts.calls.Load())
	assert.Less(t, callCount.Load(), int32(5))
}

func TestPoll_CreateAccrualError_Logged(t *testing.T) {
	accrual := 10.0
	srv := accrualServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AccrualResponse{Status: "PROCESSED", Accrual: &accrual})
	})

	os := &mockOrderService{orders: []entities.Order{{Number: "12345678903", Status: entities.StatusNew}}}
	ts := &mockTransactionService{err: errors.New("tx failed")}
	p := newTestPoller(t, srv, os, ts)

	// ошибка не возвращается — только логируется
	err := p.poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int32(1), ts.calls.Load())
}

func TestParseRetryAfter_Default(t *testing.T) {
	assert.Equal(t, 60*time.Second, parseRetryAfter("invalid"))
	assert.Equal(t, 60*time.Second, parseRetryAfter(""))
	assert.Equal(t, 60*time.Second, parseRetryAfter("-1"))
}

func TestParseRetryAfter_Valid(t *testing.T) {
	assert.Equal(t, 30*time.Second, parseRetryAfter("30"))
}
