package entities

import "time"

type TransactionType string

const (
	TransactionTypeAccrual    TransactionType = "ACCRUAL"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
)

type Transaction struct {
	ID          int64           `db:"id"`
	UserID      int64           `db:"user_id"`
	Type        TransactionType `db:"type"`
	Amount      float64         `db:"amount"`
	OrderNumber string          `db:"order_num"`
	ProcessedAt time.Time       `db:"created_at"`
}

type Balance struct {
	Current   float64
	Withdrawn float64
}

type Withdrawal struct {
	OrderNumber string
	Amount      float64
	ProcessedAt time.Time
}
