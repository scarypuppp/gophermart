package handlers

import (
	"net/http"
)

//
// GET BALANCE
//

// GetBalance Хэндлер для получения баланса пользователя
// METHOD: GET
// Content-Type: text/plain
// Возможные коды ответа:
// 200 — успешная обработка запроса.
// 401 — пользователь не авторизован.
// 500 — внутренняя ошибка сервера.
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {

}

//
// CREATE WITHDRAW
//

// CreateWithdraw Хэндлер для запроса на списание средств
// METHOD: POST
// Content-Length: 0
// Возможные коды ответа:
// 200 — успешная обработка запроса;
// 401 — пользователь не авторизован;
// 402 — на счету недостаточно средств;
// 422 — неверный номер заказа;
// 500 — внутренняя ошибка сервера.
func (h *Handler) CreateWithdraw(w http.ResponseWriter, r *http.Request) {
}

//
// GET WITHDRAWALS
//

// GetWithdrawals Хэндлер для запроса на списание средств
// METHOD: GET
// Content-Length: 0
// 200 — успешная обработка запроса;
// 204 — нет ни одного списания.
// 401 — пользователь не авторизован.
// 500 — внутренняя ошибка сервера.
func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
}
