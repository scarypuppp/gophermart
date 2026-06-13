//go:generate mockgen -destination=mocks/mock_unit_of_work.go -package=mocks . UnitOfWork
//go:generate mockgen -destination=mocks/mock_user_repository.go -package=mocks . UserRepository
//go:generate mockgen -destination=mocks/mock_order_repository.go -package=mocks . OrderRepository
package repository

import (
	"context"
	"errors"

	"github.com/scarypuppp/gophermart/internal/entities"
)

var ErrNoRows = errors.New("no rows returned")

type UnitOfWork interface {
	Users() UserRepository
	Orders() OrderRepository

	BeginTx(ctx context.Context) (UnitOfWork, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByLogin(ctx context.Context, login string) (*entities.User, error)
	CreateUser(ctx context.Context, user entities.User) (*entities.User, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order entities.Order) error
	GetOrderByNumber(ctx context.Context, number string) (*entities.Order, error)
	GetOrdersByUserId(ctx context.Context, userId int64) (*[]entities.Order, error)
}
