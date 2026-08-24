package entities

import "time"

// OrderStatus тип статуса заказа в системе лояльности.
type OrderStatus string

const (
	// StatusNew заказ загружен, но ещё не передан в accrual систему.
	StatusNew OrderStatus = "NEW"
	// StatusProcessing заказ находится в обработке accrual системой.
	StatusProcessing OrderStatus = "PROCESSING"
	// StatusInvalid заказ не принят accrual системой — начисление не будет произведено.
	StatusInvalid OrderStatus = "INVALID"
	// StatusProcessed начисление по заказу успешно рассчитано.
	StatusProcessed OrderStatus = "PROCESSED"
)

// Order представляет заказ пользователя, по которому отслеживается начисление баллов.
type Order struct {
	Number     string      `db:"number"`
	Status     OrderStatus `db:"status"`
	Accrual    *float64    `db:"accrual"`
	UserID     int64       `db:"user_id"`
	UploadedAt time.Time   `db:"uploaded_at"`
}

// ValidateOrderNumber проверяет номер заказа по алгоритму Луна.
// Возвращает true, если номер корректен.
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
