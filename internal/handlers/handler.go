package handlers

import (
	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	config       *config.Config
	logger       *zap.Logger
	userService  *service.UserService
	orderService *service.OrderService
}

func NewHandler(
	cfg *config.Config,
	logger *zap.Logger,
	userService *service.UserService,
	orderService *service.OrderService,
) *Handler {
	return &Handler{
		config:       cfg,
		logger:       logger,
		userService:  userService,
		orderService: orderService,
	}
}
