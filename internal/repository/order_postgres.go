package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/scarypuppp/gophermart/internal/entities"
)

type OrderRepositoryPostgres struct {
	exec sqlx.ExtContext
}

func NewOrderRepositoryPostgres(db *sqlx.DB) *OrderRepositoryPostgres {
	return &OrderRepositoryPostgres{exec: db}
}

func NewOrderRepositoryPostgresTx(tx *sqlx.Tx) *OrderRepositoryPostgres {
	return &OrderRepositoryPostgres{exec: tx}
}

func (r *OrderRepositoryPostgres) CreateOrder(ctx context.Context, order entities.Order) error {
	return nil
}

func (r *OrderRepositoryPostgres) GetOrderByNumber(ctx context.Context, number string) (*entities.Order, error) {
	return nil, nil
}

func (r *OrderRepositoryPostgres) GetOrdersByUserId(ctx context.Context, userId int64) (*[]entities.Order, error) {
	return nil, nil
}
