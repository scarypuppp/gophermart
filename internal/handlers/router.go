package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)
	r.Post("/api/user/orders", h.CreateOrder)
	r.Get("/api/user/orders", h.GetOrders)
	r.Get("/api/user/balance", h.GetBalance)
	r.Get("/api/user/balance/withdraw", h.CreateWithdraw)
	r.Get("/api/user/withdrawals", h.GetWithdrawals)
	return r
}
