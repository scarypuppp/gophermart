package repository

import (
	"context"

	"github.com/scarypuppp/gophermart/internal/entities"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByLogin(ctx context.Context, login string) (*entities.User, error)
	CreateUser(ctx context.Context, user entities.User) (*entities.User, error)
}
