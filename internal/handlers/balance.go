package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/scarypuppp/gophermart/internal/middlewares"
	"github.com/scarypuppp/gophermart/internal/service"
)

//
// GET BALANCE
//

type GetBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// GetBalance godoc
//
//	@Summary		Получение баланса пользователя
//	@Tags			balance
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	GetBalanceResponse
//	@Failure		401	{string}	string	"Пользователь не авторизован"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/balance [get]
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	balance, err := h.transactionService.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(GetBalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

//
// CREATE WITHDRAW
//

type CreateWithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// CreateWithdraw godoc
//
//	@Summary		Запрос на списание средств
//	@Tags			balance
//	@Accept			json
//	@Security		BearerAuth
//	@Param			body	body	CreateWithdrawRequest	true	"Данные списания"
//	@Success		200		"Успешная обработка запроса"
//	@Failure		401		{string}	string	"Пользователь не авторизован"
//	@Failure		402		{string}	string	"Недостаточно средств"
//	@Failure		422		{string}	string	"Неверный номер заказа"
//	@Failure		500		{string}	string	"Внутренняя ошибка сервера"
//	@Router			/api/user/balance/withdraw [post]
func (h *Handler) CreateWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var req CreateWithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err := h.transactionService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderNumber):
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		case errors.Is(err, service.ErrInsufficientBalance):
			http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

//
// GET WITHDRAWALS
//

type WithdrawalResponseItem struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// GetWithdrawals godoc
//
//	@Summary		Получение списка списаний
//	@Tags			balance
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		WithdrawalResponseItem
//	@Success		204	"Нет списаний"
//	@Failure		401	{string}	string	"Пользователь не авторизован"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/withdrawals [get]
func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.transactionService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	items := make([]WithdrawalResponseItem, len(withdrawals))
	for i, w := range withdrawals {
		items[i] = WithdrawalResponseItem{
			Order:       w.OrderNumber,
			Sum:         w.Amount,
			ProcessedAt: w.ProcessedAt,
		}
	}

	response, err := json.Marshal(items)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
