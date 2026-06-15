package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func ptr(f float64) *float64 { return &f }

func setupTransactionMocks(t *testing.T) (
	*gomock.Controller,
	*mocks.MockUnitOfWork,
	*mocks.MockUnitOfWork,
	*mocks.MockTransactionRepository,
) {
	ctrl := gomock.NewController(t)
	uowMock := mocks.NewMockUnitOfWork(ctrl)
	uowTxMock := mocks.NewMockUnitOfWork(ctrl)
	txRepoMock := mocks.NewMockTransactionRepository(ctrl)
	return ctrl, uowMock, uowTxMock, txRepoMock
}

func TestGetBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		_, uowMock, _, txRepoMock := setupTransactionMocks(t)

		uowMock.EXPECT().Transactions().Return(txRepoMock)
		txRepoMock.EXPECT().
			GetBalance(ctx, int64(1)).
			Return(entities.Balance{Current: 500.5, Withdrawn: 42}, nil)

		s := NewTransactionService(uowMock)
		balance, err := s.GetBalance(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, 500.5, balance.Current)
		assert.Equal(t, float64(42), balance.Withdrawn)
	})

	t.Run("repo error", func(t *testing.T) {
		_, uowMock, _, txRepoMock := setupTransactionMocks(t)
		repoErr := errors.New("db error")

		uowMock.EXPECT().Transactions().Return(txRepoMock)
		txRepoMock.EXPECT().
			GetBalance(ctx, int64(1)).
			Return(entities.Balance{}, repoErr)

		s := NewTransactionService(uowMock)
		_, err := s.GetBalance(ctx, 1)

		assert.Error(t, err)
	})
}

func TestWithdraw(t *testing.T) {
	ctx := context.Background()
	const validOrder = "12345678903"

	t.Run("success", func(t *testing.T) {
		_, uowMock, uowTxMock, txRepoMock := setupTransactionMocks(t)

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Transactions().Return(txRepoMock).Times(2)
		txRepoMock.EXPECT().
			GetBalance(ctx, int64(1)).
			Return(entities.Balance{Current: 1000, Withdrawn: 0}, nil)
		txRepoMock.EXPECT().
			CreateTransaction(ctx, entities.Transaction{
				UserID:      1,
				Amount:      -500,
				Type:        entities.TransactionTypeWithdrawal,
				OrderNumber: validOrder,
			}).
			Return(entities.Transaction{ID: 1, UserID: 1, Amount: -500, Type: entities.TransactionTypeWithdrawal, OrderNumber: validOrder}, nil)
		uowTxMock.EXPECT().Commit(ctx).Return(nil)

		s := NewTransactionService(uowMock)
		err := s.CreateWithdraw(ctx, 1, validOrder, 500)

		assert.NoError(t, err)
	})

	t.Run("invalid order number", func(t *testing.T) {
		_, uowMock, _, _ := setupTransactionMocks(t)

		s := NewTransactionService(uowMock)
		err := s.CreateWithdraw(ctx, 1, "123", 500)

		assert.ErrorIs(t, err, ErrInvalidOrderNumber)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		_, uowMock, uowTxMock, txRepoMock := setupTransactionMocks(t)

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Transactions().Return(txRepoMock)
		txRepoMock.EXPECT().
			GetBalance(ctx, int64(1)).
			Return(entities.Balance{Current: 100, Withdrawn: 0}, nil)

		s := NewTransactionService(uowMock)
		err := s.CreateWithdraw(ctx, 1, validOrder, 500)

		assert.ErrorIs(t, err, ErrInsufficientBalance)
	})

	t.Run("repo error on create transaction", func(t *testing.T) {
		_, uowMock, uowTxMock, txRepoMock := setupTransactionMocks(t)
		repoErr := errors.New("db error")

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Transactions().Return(txRepoMock).Times(2)
		txRepoMock.EXPECT().
			GetBalance(ctx, int64(1)).
			Return(entities.Balance{Current: 1000, Withdrawn: 0}, nil)
		txRepoMock.EXPECT().
			CreateTransaction(ctx, gomock.Any()).
			Return(entities.Transaction{}, repoErr)

		s := NewTransactionService(uowMock)
		err := s.CreateWithdraw(ctx, 1, validOrder, 500)

		assert.Error(t, err)
	})
}

func TestCreateAccrual(t *testing.T) {
	ctx := context.Background()
	const orderNum = "12345678903"

	t.Run("processed with accrual creates transaction", func(t *testing.T) {
		_, uowMock, uowTxMock, txRepoMock := setupTransactionMocks(t)
		ctrl := gomock.NewController(t)
		orderRepoMock := mocks.NewMockOrderRepository(ctrl)

		order := entities.Order{
			Number:  orderNum,
			UserID:  1,
			Status:  entities.StatusProcessed,
			Accrual: ptr(300),
		}

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Orders().Return(orderRepoMock)
		orderRepoMock.EXPECT().UpdateOrder(ctx, order).Return(nil)
		uowTxMock.EXPECT().Transactions().Return(txRepoMock)
		txRepoMock.EXPECT().
			CreateTransaction(ctx, entities.Transaction{
				UserID:      1,
				Amount:      300,
				Type:        entities.TransactionTypeAccrual,
				OrderNumber: orderNum,
			}).
			Return(entities.Transaction{ID: 1}, nil)
		uowTxMock.EXPECT().Commit(ctx).Return(nil)

		s := NewTransactionService(uowMock)
		err := s.CreateAccrual(ctx, order)

		assert.NoError(t, err)
	})

	t.Run("non-processed order skips transaction", func(t *testing.T) {
		_, uowMock, uowTxMock, _ := setupTransactionMocks(t)
		ctrl := gomock.NewController(t)
		orderRepoMock := mocks.NewMockOrderRepository(ctrl)

		order := entities.Order{
			Number: orderNum,
			UserID: 1,
			Status: entities.StatusProcessing,
		}

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Orders().Return(orderRepoMock)
		orderRepoMock.EXPECT().UpdateOrder(ctx, order).Return(nil)
		uowTxMock.EXPECT().Commit(ctx).Return(nil)

		s := NewTransactionService(uowMock)
		err := s.CreateAccrual(ctx, order)

		assert.NoError(t, err)
	})

	t.Run("update orders error", func(t *testing.T) {
		_, uowMock, uowTxMock, _ := setupTransactionMocks(t)
		ctrl := gomock.NewController(t)
		orderRepoMock := mocks.NewMockOrderRepository(ctrl)
		repoErr := errors.New("db error")

		order := entities.Order{Number: orderNum, UserID: 1, Status: entities.StatusProcessed, Accrual: ptr(100)}

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Orders().Return(orderRepoMock)
		orderRepoMock.EXPECT().UpdateOrder(ctx, order).Return(repoErr)

		s := NewTransactionService(uowMock)
		err := s.CreateAccrual(ctx, order)

		assert.Error(t, err)
	})

	t.Run("create transaction error", func(t *testing.T) {
		_, uowMock, uowTxMock, txRepoMock := setupTransactionMocks(t)
		ctrl := gomock.NewController(t)
		orderRepoMock := mocks.NewMockOrderRepository(ctrl)
		repoErr := errors.New("db error")

		order := entities.Order{Number: orderNum, UserID: 1, Status: entities.StatusProcessed, Accrual: ptr(100)}

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Orders().Return(orderRepoMock)
		orderRepoMock.EXPECT().UpdateOrder(ctx, order).Return(nil)
		uowTxMock.EXPECT().Transactions().Return(txRepoMock)
		txRepoMock.EXPECT().CreateTransaction(ctx, gomock.Any()).Return(entities.Transaction{}, repoErr)

		s := NewTransactionService(uowMock)
		err := s.CreateAccrual(ctx, order)

		assert.Error(t, err)
	})
}

func TestGetWithdrawals(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		_, uowMock, _, txRepoMock := setupTransactionMocks(t)

		expected := []entities.Withdrawal{
			{OrderNumber: "12345678903", Amount: 500, ProcessedAt: time.Now()},
			{OrderNumber: "49927398716", Amount: 100, ProcessedAt: time.Now().Add(-time.Hour)},
		}

		uowMock.EXPECT().Transactions().Return(txRepoMock)
		txRepoMock.EXPECT().
			GetWithdrawals(ctx, int64(1)).
			Return(expected, nil)

		s := NewTransactionService(uowMock)
		withdrawals, err := s.GetWithdrawals(ctx, 1)

		assert.NoError(t, err)
		assert.Len(t, withdrawals, 2)
		assert.Equal(t, expected[0].OrderNumber, withdrawals[0].OrderNumber)
		assert.Equal(t, expected[0].Amount, withdrawals[0].Amount)
	})

	t.Run("empty list", func(t *testing.T) {
		_, uowMock, _, txRepoMock := setupTransactionMocks(t)

		uowMock.EXPECT().Transactions().Return(txRepoMock)
		txRepoMock.EXPECT().
			GetWithdrawals(ctx, int64(1)).
			Return([]entities.Withdrawal{}, nil)

		s := NewTransactionService(uowMock)
		withdrawals, err := s.GetWithdrawals(ctx, 1)

		assert.NoError(t, err)
		assert.Len(t, withdrawals, 0)
	})

	t.Run("repo error", func(t *testing.T) {
		_, uowMock, _, txRepoMock := setupTransactionMocks(t)
		repoErr := errors.New("db error")

		uowMock.EXPECT().Transactions().Return(txRepoMock)
		txRepoMock.EXPECT().
			GetWithdrawals(ctx, int64(1)).
			Return(nil, repoErr)

		s := NewTransactionService(uowMock)
		_, err := s.GetWithdrawals(ctx, 1)

		assert.Error(t, err)
	})
}
