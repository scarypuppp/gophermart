package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/scarypuppp/gophermart/internal/middlewares"
	"github.com/scarypuppp/gophermart/internal/service"
)

//
// CREATE ORDER
//

// CreateOrder godoc
//
//	@Summary		Загрузка номера заказа
//	@Tags			orders
//	@Accept			plain
//	@Produce		plain
//	@Security		BearerAuth
//	@Param			body	body		string	true	"Номер заказа"
//	@Success		200		"Заказ уже был загружен этим пользователем"
//	@Success		202		"Новый заказ принят в обработку"
//	@Failure		400		{string}	string	"Неверный формат запроса"
//	@Failure		401		{string}	string	"Пользователь не аутентифицирован"
//	@Failure		409		{string}	string	"Заказ уже загружен другим пользователем"
//	@Failure		422		{string}	string	"Неверный формат номера заказа"
//	@Failure		500		{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/orders [post]
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	var buff bytes.Buffer
	_, err := buff.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, created, err := h.orderService.GetOrCreateOrder(r.Context(), userID, buff.String())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderNumber):
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		case errors.Is(err, service.ErrOrderAssociatedWithOtherUser):
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	if created {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	w.WriteHeader(http.StatusOK)
}

//
// GET ORDERS
//

// GetOrdersResponseItem описывает элемент ответа хэндлера GetOrders
type GetOrdersResponseItem struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// GetOrdersResponse список элементов хэндлера GetOrders
type GetOrdersResponse []GetOrdersResponseItem

// GetOrders godoc
//
//	@Summary		Получение списка заказов
//	@Tags			orders
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		GetOrdersResponseItem
//	@Success		204	"Нет данных"
//	@Failure		401	{string}	string	"Пользователь не авторизован"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/orders [get]
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
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var responseItems GetOrdersResponse
	for _, order := range orders {
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
