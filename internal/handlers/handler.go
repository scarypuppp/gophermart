package handlers

import (
	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/service"
)

type Handler struct {
	config      *config.Config
	userService *service.UserService
}

func NewHandler(cfg *config.Config, us *service.UserService) *Handler {
	return &Handler{
		config:      cfg,
		userService: us,
	}
}
