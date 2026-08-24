package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/repository"
)

// ErrInvalidOrderNumber возвращается, если номер заказа не прошёл проверку по алгоритму Луна.
var ErrInvalidOrderNumber = errors.New("invalid order number provided")

// ErrOrderAssociatedWithOtherUser возвращается, если заказ уже зарегистрирован другим пользователем.
var ErrOrderAssociatedWithOtherUser = errors.New("order number is  already associated with other user")

// OrderService реализует бизнес-логику управления заказами.
type OrderService struct {
	uow repository.UnitOfWork
}

// NewOrderService создаёт новый OrderService с переданным UnitOfWork.
func NewOrderService(uow repository.UnitOfWork) *OrderService {
	return &OrderService{uow}
}

// GetOrCreateOrder возвращает существующий заказ по номеру или создаёт новый для указанного пользователя.
// Второй возвращаемый параметр — true, если заказ был создан в этом вызове.
// Возвращает ErrInvalidOrderNumber при некорректном номере и ErrOrderAssociatedWithOtherUser,
// если заказ уже принадлежит другому пользователю.
func (s *OrderService) GetOrCreateOrder(ctx context.Context, userId int64, number string) (*entities.Order, bool, error) {
	if ok := entities.ValidateOrderNumber(number); ok != true {
		return nil, false, ErrInvalidOrderNumber
	}
	tx, err := s.uow.BeginTx(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("CreateOrder: error creating Tx: %w", err)
	}
	defer tx.Rollback(ctx)
	existingOrder, err := tx.Orders().GetOrderByNumber(ctx, number)
	if err == nil {
		if existingOrder.UserID != userId {
			return nil, false, ErrOrderAssociatedWithOtherUser
		}
		return existingOrder, false, nil
	}

	order := entities.Order{
		Number:     number,
		Status:     entities.StatusNew,
		UploadedAt: time.Now(),
		UserID:     userId,
	}
	err = tx.Orders().CreateOrder(ctx, order)
	if err != nil {
		return nil, false, fmt.Errorf("CreateOrder: error creating order: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("CreateOrder: error commit order: %w", err)
	}
	return &order, true, nil
}

// GetOrders возвращает все заказы пользователя, отсортированные по дате загрузки.
func (s *OrderService) GetOrders(ctx context.Context, userId int64) ([]entities.Order, error) {
	orders, err := s.uow.Orders().GetOrdersByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("GetOrders: error getting orders: %w", err)
	}
	return orders, nil
}

// GetOrdersToPoll возвращает заказы в статусах NEW и PROCESSING, ожидающие опроса accrual системы.
func (s *OrderService) GetOrdersToPoll(ctx context.Context) ([]entities.Order, error) {
	orders, err := s.uow.Orders().GetUnprocessedOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetOrdersToPoll: %w", err)
	}
	return orders, nil
}
