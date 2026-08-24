package handlers

import (
	"context"

	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/entities"
	"go.uber.org/zap"
)

type userService interface {
	RegisterUser(ctx context.Context, login, password string) (*entities.User, error)
	LoginUser(ctx context.Context, login, password string) (*entities.User, error)
}

type orderService interface {
	GetOrCreateOrder(ctx context.Context, userID int64, number string) (*entities.Order, bool, error)
	GetOrders(ctx context.Context, userID int64) ([]entities.Order, error)
}

type transactionService interface {
	GetBalance(ctx context.Context, userID int64) (entities.Balance, error)
	CreateWithdraw(ctx context.Context, userID int64, orderNumber string, amount float64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]entities.Withdrawal, error)
}

type Handler struct {
	config             *config.Config
	logger             *zap.Logger
	userService        userService
	orderService       orderService
	transactionService transactionService
}

func NewHandler(
	cfg *config.Config,
	logger *zap.Logger,
	userService userService,
	orderService orderService,
	transactionService transactionService,
) *Handler {
	return &Handler{
		config:             cfg,
		logger:             logger,
		userService:        userService,
		orderService:       orderService,
		transactionService: transactionService,
	}
}
