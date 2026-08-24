package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/repository"
)

// ErrInsufficientBalance возвращается при попытке списания суммы, превышающей текущий баланс.
var ErrInsufficientBalance = errors.New("insufficient balance")

// TransactionService реализует бизнес-логику операций с балансом пользователя.
type TransactionService struct {
	uow repository.UnitOfWork
}

// NewTransactionService создаёт новый TransactionService с переданным UnitOfWork.
func NewTransactionService(uow repository.UnitOfWork) *TransactionService {
	return &TransactionService{uow}
}

// GetBalance возвращает текущий баланс и суммарную сумму списаний для пользователя.
func (s *TransactionService) GetBalance(ctx context.Context, userID int64) (entities.Balance, error) {
	balance, err := s.uow.Transactions().GetBalance(ctx, userID)
	if err != nil {
		return entities.Balance{}, fmt.Errorf("GetBalance: %w", err)
	}
	return balance, nil
}

// CreateWithdraw списывает amount баллов с баланса пользователя в счёт оплаты заказа.
// Возвращает ErrInvalidOrderNumber при некорректном номере заказа и ErrInsufficientBalance,
// если баланс не покрывает запрошенную сумму.
func (s *TransactionService) CreateWithdraw(ctx context.Context, userID int64, orderNumber string, amount float64) error {
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

// CreateAccrual обновляет статус заказа и, если он перешёл в PROCESSED, начисляет баллы на баланс пользователя.
func (s *TransactionService) CreateAccrual(ctx context.Context, order entities.Order) error {
	tx, err := s.uow.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("CreateAccrual: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := tx.Orders().UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("CreateAccrual: UpdateOrder: %w", err)
	}

	if order.Status == entities.StatusProcessed && order.Accrual != nil {
		_, err = tx.Transactions().CreateTransaction(ctx, entities.Transaction{
			UserID:      order.UserID,
			Amount:      *order.Accrual,
			Type:        entities.TransactionTypeAccrual,
			OrderNumber: order.Number,
		})
		if err != nil {
			return fmt.Errorf("CreateAccrual: CreateTransaction: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetWithdrawals возвращает историю списаний пользователя, отсортированную по дате.
func (s *TransactionService) GetWithdrawals(ctx context.Context, userID int64) ([]entities.Withdrawal, error) {
	withdrawals, err := s.uow.Transactions().GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetWithdrawals: %w", err)
	}
	return withdrawals, nil
}
