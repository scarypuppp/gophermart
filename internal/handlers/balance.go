package handlers

import (
	"net/http"
)

//
// GET BALANCE
//

// GetBalance godoc
//
//	@Summary		Получение баланса пользователя
//	@Tags			balance
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	"Успешная обработка запроса"
//	@Failure		401	{string}	string	"Пользователь не авторизован"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/balance [get]
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {

}

//
// CREATE WITHDRAW
//

// CreateWithdraw godoc
//
//	@Summary		Запрос на списание средств
//	@Tags			balance
//	@Security		BearerAuth
//	@Success		200	"Успешная обработка запроса"
//	@Failure		401	{string}	string	"Пользователь не авторизован"
//	@Failure		402	{string}	string	"Недостаточно средств"
//	@Failure		422	{string}	string	"Неверный номер заказа"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/balance/withdraw [get]
func (h *Handler) CreateWithdraw(w http.ResponseWriter, r *http.Request) {
}

//
// GET WITHDRAWALS
//

// GetWithdrawals godoc
//
//	@Summary		Получение списка списаний
//	@Tags			balance
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	"Успешная обработка запроса"
//	@Success		204	"Нет списаний"
//	@Failure		401	{string}	string	"Пользователь не авторизован"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/withdrawals [get]
func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
}
