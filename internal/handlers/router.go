package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)
	return r
}
