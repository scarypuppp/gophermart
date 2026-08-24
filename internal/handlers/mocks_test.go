package handlers

import (
	"context"
	"net/http"

	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/middlewares"
	"go.uber.org/zap"
)

type mockUserService struct {
	registerFn func(ctx context.Context, login, password string) (*entities.User, error)
	loginFn    func(ctx context.Context, login, password string) (*entities.User, error)
}

func (m *mockUserService) RegisterUser(ctx context.Context, login, password string) (*entities.User, error) {
	return m.registerFn(ctx, login, password)
}
func (m *mockUserService) LoginUser(ctx context.Context, login, password string) (*entities.User, error) {
	return m.loginFn(ctx, login, password)
}

type mockOrderService struct {
	getOrCreateFn func(ctx context.Context, userID int64, number string) (*entities.Order, bool, error)
	getOrdersFn   func(ctx context.Context, userID int64) ([]entities.Order, error)
}

func (m *mockOrderService) GetOrCreateOrder(ctx context.Context, userID int64, number string) (*entities.Order, bool, error) {
	return m.getOrCreateFn(ctx, userID, number)
}
func (m *mockOrderService) GetOrders(ctx context.Context, userID int64) ([]entities.Order, error) {
	return m.getOrdersFn(ctx, userID)
}

type mockTransactionService struct {
	getBalanceFn     func(ctx context.Context, userID int64) (entities.Balance, error)
	createWithdrawFn func(ctx context.Context, userID int64, order string, sum float64) error
	getWithdrawalsFn func(ctx context.Context, userID int64) ([]entities.Withdrawal, error)
}

func (m *mockTransactionService) GetBalance(ctx context.Context, userID int64) (entities.Balance, error) {
	return m.getBalanceFn(ctx, userID)
}
func (m *mockTransactionService) CreateWithdraw(ctx context.Context, userID int64, order string, sum float64) error {
	return m.createWithdrawFn(ctx, userID, order, sum)
}
func (m *mockTransactionService) GetWithdrawals(ctx context.Context, userID int64) ([]entities.Withdrawal, error) {
	return m.getWithdrawalsFn(ctx, userID)
}

func testHandler(us userService, os orderService, ts transactionService) *Handler {
	cfg := &config.Config{SecretKey: "test", TokenExpSeconds: 3600}
	return NewHandler(cfg, zap.NewNop(), us, os, ts)
}

func withUserID(r *http.Request, id int64) *http.Request {
	ctx := context.WithValue(r.Context(), middlewares.UserIDKey, id)
	return r.WithContext(ctx)
}
