package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBalance_Success(t *testing.T) {
	ts := &mockTransactionService{
		getBalanceFn: func(_ context.Context, _ int64) (entities.Balance, error) {
			return entities.Balance{Current: 100.0, Withdrawn: 20.0}, nil
		},
	}
	h := testHandler(nil, nil, ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/user/balance", nil), 1)
	h.GetBalance(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp GetBalanceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 100.0, resp.Current)
	assert.Equal(t, 20.0, resp.Withdrawn)
}

func TestCreateWithdraw_Success(t *testing.T) {
	ts := &mockTransactionService{
		createWithdrawFn: func(_ context.Context, _ int64, _ string, _ float64) error {
			return nil
		},
	}
	h := testHandler(nil, nil, ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(`{"order":"12345678903","sum":50.0}`)), 1)
	h.CreateWithdraw(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateWithdraw_InsufficientBalance(t *testing.T) {
	ts := &mockTransactionService{
		createWithdrawFn: func(_ context.Context, _ int64, _ string, _ float64) error {
			return service.ErrInsufficientBalance
		},
	}
	h := testHandler(nil, nil, ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(`{"order":"12345678903","sum":9999.0}`)), 1)
	h.CreateWithdraw(w, r)

	assert.Equal(t, http.StatusPaymentRequired, w.Code)
}

func TestCreateWithdraw_InvalidOrder(t *testing.T) {
	ts := &mockTransactionService{
		createWithdrawFn: func(_ context.Context, _ int64, _ string, _ float64) error {
			return service.ErrInvalidOrderNumber
		},
	}
	h := testHandler(nil, nil, ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(`{"order":"bad","sum":10.0}`)), 1)
	h.CreateWithdraw(w, r)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestGetWithdrawals_WithData(t *testing.T) {
	ts := &mockTransactionService{
		getWithdrawalsFn: func(_ context.Context, _ int64) ([]entities.Withdrawal, error) {
			return []entities.Withdrawal{
				{OrderNumber: "12345678903", Amount: 10.0, ProcessedAt: time.Now()},
			}, nil
		},
	}
	h := testHandler(nil, nil, ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil), 1)
	h.GetWithdrawals(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var items []WithdrawalResponseItem
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	assert.Len(t, items, 1)
}

func TestGetWithdrawals_Empty(t *testing.T) {
	ts := &mockTransactionService{
		getWithdrawalsFn: func(_ context.Context, _ int64) ([]entities.Withdrawal, error) {
			return nil, nil
		},
	}
	h := testHandler(nil, nil, ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil), 1)
	h.GetWithdrawals(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGetWithdrawals_ServiceError(t *testing.T) {
	ts := &mockTransactionService{
		getWithdrawalsFn: func(_ context.Context, _ int64) ([]entities.Withdrawal, error) {
			return nil, errors.New("db error")
		},
	}
	h := testHandler(nil, nil, ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil), 1)
	h.GetWithdrawals(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
