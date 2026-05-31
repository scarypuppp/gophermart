package repository

import (
	"context"

	"github.com/scarypuppp/gophermart/internal/entities"
)

type UserPostgresRepository struct{}

func NewUserPostgresRepository() *UserPostgresRepository {
	return &UserPostgresRepository{}
}

func (r *UserPostgresRepository) GetById(ctx context.Context, id int64) (*entities.User, error) {
	return nil, nil
}

func (r *UserPostgresRepository) GetByLogin(ctx context.Context, login string) (*entities.User, error) {
	return nil, nil
}
func (r *UserPostgresRepository) CreateUser(ctx context.Context, user entities.User) (*entities.User, error) {
	return nil, nil
}
