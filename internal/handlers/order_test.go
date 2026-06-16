package handlers

import (
	"context"
	"encoding/json"
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

func TestCreateOrder_NewOrder(t *testing.T) {
	os := &mockOrderService{
		getOrCreateFn: func(_ context.Context, _ int64, _ string) (*entities.Order, bool, error) {
			return &entities.Order{}, true, nil
		},
	}
	h := testHandler(nil, os, nil)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903")), 1)
	h.CreateOrder(w, r)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestCreateOrder_AlreadyOwned(t *testing.T) {
	os := &mockOrderService{
		getOrCreateFn: func(_ context.Context, _ int64, _ string) (*entities.Order, bool, error) {
			return &entities.Order{}, false, nil
		},
	}
	h := testHandler(nil, os, nil)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903")), 1)
	h.CreateOrder(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateOrder_Conflict(t *testing.T) {
	os := &mockOrderService{
		getOrCreateFn: func(_ context.Context, _ int64, _ string) (*entities.Order, bool, error) {
			return nil, false, service.ErrOrderAssociatedWithOtherUser
		},
	}
	h := testHandler(nil, os, nil)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903")), 1)
	h.CreateOrder(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateOrder_InvalidNumber(t *testing.T) {
	os := &mockOrderService{
		getOrCreateFn: func(_ context.Context, _ int64, _ string) (*entities.Order, bool, error) {
			return nil, false, service.ErrInvalidOrderNumber
		},
	}
	h := testHandler(nil, os, nil)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("bad")), 1)
	h.CreateOrder(w, r)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateOrder_NoUserID(t *testing.T) {
	h := testHandler(nil, nil, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	h.CreateOrder(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetOrders_WithData(t *testing.T) {
	accrual := 10.5
	os := &mockOrderService{
		getOrdersFn: func(_ context.Context, _ int64) ([]entities.Order, error) {
			return []entities.Order{
				{Number: "12345678903", Status: entities.StatusProcessed, Accrual: &accrual, UploadedAt: time.Now()},
			}, nil
		},
	}
	h := testHandler(nil, os, nil)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/user/orders", nil), 1)
	h.GetOrders(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var items []GetOrdersResponseItem
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	assert.Len(t, items, 1)
	assert.Equal(t, "12345678903", items[0].Number)
}

func TestGetOrders_Empty(t *testing.T) {
	os := &mockOrderService{
		getOrdersFn: func(_ context.Context, _ int64) ([]entities.Order, error) {
			return nil, nil
		},
	}
	h := testHandler(nil, os, nil)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/user/orders", nil), 1)
	h.GetOrders(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
