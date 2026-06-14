package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/scarypuppp/gophermart/internal/middlewares"
	"github.com/scarypuppp/gophermart/internal/service"
)

//
// CREATE ORDER
//

// CreateOrder Хэндлер для создания заказа
// METHOD: POST
// Content-Type: text/plain
// Возможные коды ответа:
// 200 — номер заказа уже был загружен этим пользователем;
// 202 — новый номер заказа принят в обработку;
// 400 — неверный формат запроса;
// 401 — пользователь не аутентифицирован;
// 409 — номер заказа уже был загружен другим пользователем;
// 422 — неверный формат номера заказа;
// 500 — внутренняя ошибка сервера.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var buff bytes.Buffer
	_, err := buff.ReadFrom(r.Body)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, created, err := h.orderService.GetOrCreateOrder(r.Context(), 1, buff.String())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderNumber):
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		case errors.Is(err, service.ErrOrderAssociatedWithOtherUser):
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		default:
			fmt.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	if !created {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

//
// GET ORDERS
//

// GetOrdersResponseItem описывает элемент ответа хэндлера GetOrders
type GetOrdersResponseItem struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *int64    `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// GetOrdersResponse список элементов хэндлера GetOrders
type GetOrdersResponse []GetOrdersResponseItem

// GetOrders Хэндлер для получения заказов
// METHOD: GET
// Content-Length: 0
// Возможные коды ответа:
// 200 — успешная обработка запроса.
// 204 — нет данных для ответа.
// 401 — пользователь не авторизован.
// 500 — внутренняя ошибка сервера.
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	orders, err := h.orderService.GetOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(*orders) == 0 {
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
		return
	}

	var responseItems GetOrdersResponse
	for _, order := range *orders {
		responseItems = append(responseItems, GetOrdersResponseItem{
			Number:     order.Number,
			Status:     string(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	response, err := json.Marshal(responseItems)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
