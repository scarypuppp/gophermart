package entities

import "time"

// TransactionType тип операции с балансом пользователя.
type TransactionType string

const (
	// TransactionTypeAccrual начисление баллов по заказу.
	TransactionTypeAccrual TransactionType = "ACCRUAL"
	// TransactionTypeWithdrawal списание баллов при оплате заказа.
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
)

// Transaction представляет одну операцию изменения баланса пользователя.
type Transaction struct {
	ID          int64           `db:"id"`
	UserID      int64           `db:"user_id"`
	Type        TransactionType `db:"type"`
	Amount      float64         `db:"amount"`
	OrderNumber string          `db:"order_num"`
	ProcessedAt time.Time       `db:"created_at"`
}

// Balance текущее состояние баланса пользователя.
type Balance struct {
	Current   float64
	Withdrawn float64
}

// Withdrawal представляет запись о списании баллов при оплате заказа.
type Withdrawal struct {
	OrderNumber string
	Amount      float64
	ProcessedAt time.Time
}
