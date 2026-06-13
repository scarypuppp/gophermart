package service

import (
	"context"
	"testing"
	"time"

	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/repository"
	"github.com/scarypuppp/gophermart/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupOrderMocks(t *testing.T) (
	*gomock.Controller,
	*mocks.MockUnitOfWork,
	*mocks.MockUnitOfWork,
	*mocks.MockOrderRepository,
) {
	ctrl := gomock.NewController(t)
	uowMock := mocks.NewMockUnitOfWork(ctrl)
	uowTxMock := mocks.NewMockUnitOfWork(ctrl)
	orderRepoMock := mocks.NewMockOrderRepository(ctrl)
	return ctrl, uowMock, uowTxMock, orderRepoMock
}

func TestCreateOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("success create", func(t *testing.T) {
		_, uowMock, uowTxMock, orderRepoMock := setupOrderMocks(t)

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Orders().Return(orderRepoMock).Times(2)
		orderRepoMock.EXPECT().
			GetOrderByNumber(ctx, "12345678903").
			Return(nil, repository.ErrNoRows)
		orderRepoMock.EXPECT().
			CreateOrder(ctx, gomock.Any()).
			Return(nil)
		uowTxMock.EXPECT().Commit(ctx).Return(nil)

		s := NewOrderService(uowMock)
		order, created, err := s.GetOrCreateOrder(ctx, 1, "12345678903")
		assert.NoError(t, err)
		assert.Equal(t, created, true)
		assert.NotNil(t, order)
	})

	t.Run("success get", func(t *testing.T) {
		_, uowMock, uowTxMock, orderRepoMock := setupOrderMocks(t)

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Orders().Return(orderRepoMock)
		orderRepoMock.EXPECT().
			GetOrderByNumber(ctx, "12345678903").
			Return(
				&entities.Order{
					Number:     "12345678903",
					Status:     entities.StatusProcessing,
					UserID:     1,
					UploadedAt: time.Now().Add(-time.Second * 10),
				},
				nil,
			)

		s := NewOrderService(uowMock)
		order, created, err := s.GetOrCreateOrder(ctx, 1, "12345678903")
		assert.NoError(t, err)
		assert.Equal(t, created, false)
		assert.NotNil(t, order)
	})

	t.Run("order number invalid", func(t *testing.T) {
		_, uowMock, _, _ := setupOrderMocks(t)

		s := NewOrderService(uowMock)
		_, _, err := s.GetOrCreateOrder(ctx, 1, "123")
		assert.ErrorIs(t, err, ErrInvalidOrderNumber)
	})

	t.Run("order number associated with other user", func(t *testing.T) {
		_, uowMock, uowTxMock, orderRepoMock := setupOrderMocks(t)
		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Orders().Return(orderRepoMock)
		orderRepoMock.EXPECT().
			GetOrderByNumber(ctx, "12345678903").
			Return(
				&entities.Order{
					Number:     "12345678903",
					Status:     entities.StatusProcessing,
					UserID:     2,
					UploadedAt: time.Now().Add(-time.Second * 10),
				},
				nil,
			)

		s := NewOrderService(uowMock)
		_, _, err := s.GetOrCreateOrder(ctx, 1, "12345678903")
		assert.ErrorIs(t, err, ErrOrderAssociatedWithOtherUser)
	})
}

func TestGetOrders(t *testing.T) {
	ctx := context.Background()

	t.Run("success get", func(t *testing.T) {
		_, uowMock, _, orderRepoMock := setupOrderMocks(t)
		uowMock.EXPECT().Orders().Return(orderRepoMock)
		orderRepoMock.EXPECT().GetOrdersByUserId(ctx, gomock.Any()).
			Return(&[]entities.Order{{Number: "123", Status: entities.StatusNew, UserID: 1}}, nil)

		s := NewOrderService(uowMock)
		orders, err := s.GetOrders(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, orders)
		assert.Len(t, *orders, 1)
	})
	t.Run("empty list on no rows", func(t *testing.T) {
		_, uowMock, _, orderRepoMock := setupOrderMocks(t)
		uowMock.EXPECT().Orders().Return(orderRepoMock)
		orderRepoMock.EXPECT().GetOrdersByUserId(ctx, gomock.Any()).
			Return(nil, repository.ErrNoRows)

		s := NewOrderService(uowMock)
		orders, err := s.GetOrders(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, orders)
		assert.Len(t, *orders, 0)
	})
}
