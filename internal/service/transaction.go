package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/repository"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

type TransactionService struct {
	uow repository.UnitOfWork
}

func NewTransactionService(uow repository.UnitOfWork) *TransactionService {
	return &TransactionService{uow}
}

func (s *TransactionService) GetBalance(ctx context.Context, userID int64) (entities.Balance, error) {
	balance, err := s.uow.Transactions().GetBalance(ctx, userID)
	if err != nil {
		return entities.Balance{}, fmt.Errorf("GetBalance: %w", err)
	}
	return balance, nil
}

func (s *TransactionService) Withdraw(ctx context.Context, userID int64, orderNumber string, amount float64) error {
	if ok := entities.ValidateOrderNumber(orderNumber); !ok {
		return ErrInvalidOrderNumber
	}

	tx, err := s.uow.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("Withdraw: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)

	balance, err := tx.Transactions().GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("Withdraw: GetBalance: %w", err)
	}
	if balance.Current < amount {
		return ErrInsufficientBalance
	}

	_, err = tx.Transactions().CreateTransaction(ctx, entities.Transaction{
		UserID:      userID,
		Amount:      -amount,
		Type:        entities.TransactionTypeWithdrawal,
		OrderNumber: orderNumber,
	})
	if err != nil {
		return fmt.Errorf("Withdraw: CreateTransaction: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *TransactionService) GetWithdrawals(ctx context.Context, userID int64) ([]entities.Withdrawal, error) {
	withdrawals, err := s.uow.Transactions().GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetWithdrawals: %w", err)
	}
	return withdrawals, nil
}
