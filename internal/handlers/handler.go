package handlers

import "github.com/scarypuppp/gophermart/internal/service"

type Handler struct {
	userService *service.UserService
}

func NewHandler(us *service.UserService) *Handler {
	return &Handler{
		userService: us,
	}
}
