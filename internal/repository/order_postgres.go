package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	const insertQuery = `
    INSERT INTO orders (number, status, accrual, user_id, uploaded_at)
    VALUES (:number, :status, :accrual, :user_id, :uploaded_at)`
	_, err := sqlx.NamedExecContext(ctx, r.exec, insertQuery, order)
	if err != nil {
		return fmt.Errorf("CreateOrder: %w", err)
	}
	return nil
}

func (r *OrderRepositoryPostgres) GetOrderByNumber(ctx context.Context, number string) (*entities.Order, error) {
	var order entities.Order
	err := sqlx.GetContext(ctx, r.exec, &order,
		`SELECT number, status, accrual, user_id, uploaded_at FROM orders WHERE number = $1`, number,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("GetOrderByNumber: %w", err)
	}
	return &order, nil
}

func (r *OrderRepositoryPostgres) GetOrdersByUserId(ctx context.Context, userId int64) ([]entities.Order, error) {
	var orders []entities.Order
	err := sqlx.SelectContext(ctx, r.exec, &orders,
		`SELECT number, status, accrual, user_id, uploaded_at FROM orders WHERE user_id = $1`, userId,
	)
	if err != nil {
		return nil, fmt.Errorf("GetOrdersByUserId: %w", err)
	}
	return orders, nil
}
