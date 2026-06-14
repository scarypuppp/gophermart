package entities

import "time"

type OrderStatus string

const (
	StatusNew        OrderStatus = "NEW"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number     string      `db:"number"`
	Status     OrderStatus `db:"status"`
	Accrual    *float64    `db:"accrual"`
	UserID     int64       `db:"user_id"`
	UploadedAt time.Time   `db:"uploaded_at"`
}

func ValidateOrderNumber(number string) bool {
	if len(number) == 0 {
		return false
	}

	sum := 0
	nDigits := len(number)
	parity := nDigits % 2

	for i := 0; i < nDigits; i++ {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
