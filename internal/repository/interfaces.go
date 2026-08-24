//go:generate mockgen -destination=mocks/mock_unit_of_work.go -package=mocks . UnitOfWork
//go:generate mockgen -destination=mocks/mock_user_repository.go -package=mocks . UserRepository
//go:generate mockgen -destination=mocks/mock_order_repository.go -package=mocks . OrderRepository
//go:generate mockgen -destination=mocks/mock_transaction_repository.go -package=mocks . TransactionRepository
package repository

import (
	"context"
	"errors"

	"github.com/scarypuppp/gophermart/internal/entities"
)

// ErrNoRows возвращается репозиторием, когда запрос не вернул ни одной строки.
var ErrNoRows = errors.New("no rows returned")

// UnitOfWork объединяет все репозитории и управляет транзакциями базы данных.
// BeginTx открывает новую транзакцию и возвращает UnitOfWork, работающий внутри неё.
type UnitOfWork interface {
	Users() UserRepository
	Orders() OrderRepository
	Transactions() TransactionRepository

	BeginTx(ctx context.Context) (UnitOfWork, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// UserRepository интерфейс для операций с пользователями в хранилище.
type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByLogin(ctx context.Context, login string) (*entities.User, error)
	CreateUser(ctx context.Context, user entities.User) (*entities.User, error)
}

// OrderRepository интерфейс для операций с заказами в хранилище.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order entities.Order) error
	GetOrderByNumber(ctx context.Context, number string) (*entities.Order, error)
	GetOrdersByUserId(ctx context.Context, userId int64) ([]entities.Order, error)
	GetUnprocessedOrders(ctx context.Context) ([]entities.Order, error)
	UpdateOrder(ctx context.Context, order entities.Order) error
}

// TransactionRepository интерфейс для операций с транзакциями баланса в хранилище.
type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction entities.Transaction) (entities.Transaction, error)
	GetBalance(ctx context.Context, userId int64) (entities.Balance, error)
	GetWithdrawals(ctx context.Context, userId int64) ([]entities.Withdrawal, error)
}
