package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/scarypuppp/gophermart/internal/entities"
)

type TransactionRepositoryPostgres struct {
	exec sqlx.ExtContext
}

func NewTransactionRepositoryPostgres(db *sqlx.DB) *TransactionRepositoryPostgres {
	return &TransactionRepositoryPostgres{exec: db}
}

func NewTransactionRepositoryPostgresTx(tx *sqlx.Tx) *TransactionRepositoryPostgres {
	return &TransactionRepositoryPostgres{exec: tx}
}

func (r *TransactionRepositoryPostgres) CreateTransaction(ctx context.Context, tx entities.Transaction) (entities.Transaction, error) {
	const query = `
		INSERT INTO transactions (user_id, type, amount, order_num)
		VALUES (:user_id, :type, :amount, :order_num)
		RETURNING id, created_at`
	rows, err := sqlx.NamedQueryContext(ctx, r.exec, query, tx)
	if err != nil {
		return entities.Transaction{}, fmt.Errorf("CreateTransaction: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&tx.ID, &tx.ProcessedAt); err != nil {
			return entities.Transaction{}, fmt.Errorf("CreateTransaction scan: %w", err)
		}
	}
	return tx, nil
}

func (r *TransactionRepositoryPostgres) GetBalance(ctx context.Context, userID int64) (entities.Balance, error) {
	var row struct {
		Current   float64 `db:"current"`
		Withdrawn float64 `db:"withdrawn"`
	}
	const query = `
		SELECT
			COALESCE(SUM(amount), 0)       AS current,
			COALESCE(ABS(SUM(amount) FILTER (WHERE amount < 0)), 0)  AS withdrawn
		FROM transactions
		WHERE user_id = $1`
	if err := sqlx.GetContext(ctx, r.exec, &row, query, userID); err != nil {
		return entities.Balance{}, fmt.Errorf("GetBalance: %w", err)
	}
	return entities.Balance{Current: row.Current, Withdrawn: row.Withdrawn}, nil
}

func (r *TransactionRepositoryPostgres) GetWithdrawals(ctx context.Context, userID int64) ([]entities.Withdrawal, error) {
	var txs []entities.Transaction
	const query = `
		SELECT order_num, amount, created_at
		FROM transactions
		WHERE user_id = $1 AND amount < 0
		ORDER BY created_at DESC`
	if err := sqlx.SelectContext(ctx, r.exec, &txs, query, userID); err != nil {
		return nil, fmt.Errorf("GetWithdrawals: %w", err)
	}

	result := make([]entities.Withdrawal, len(txs))
	for i, tx := range txs {
		result[i] = entities.Withdrawal{
			OrderNumber: tx.OrderNumber,
			Amount:      -tx.Amount,
			ProcessedAt: tx.ProcessedAt,
		}
	}
	return result, nil
}
