package accrual_poller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

// ErrNotRegistered ошибка возвращается, если accrual система не зарегистрировала номер заказа.
var ErrNotRegistered = errors.New("order not registered in accrual")

// ErrTooManyRequests ошибка возвращается, если превышен лимит запросов в accrual систему.
var ErrTooManyRequests = errors.New("accrual rate limit exceeded")

// AccrualResponse описывает тело ответа от accrual системы на запрос статуса заказа.
type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

// RetryAfterError возвращается при HTTP 429 и содержит время ожидания до следующего запроса.
type RetryAfterError struct {
	RetryAfter time.Duration
}

// Error реализует интерфейс error.
func (e *RetryAfterError) Error() string {
	return fmt.Sprintf("%s: retry after %s", ErrTooManyRequests, e.RetryAfter)
}

// Unwrap позволяет сопоставить RetryAfterError с ErrTooManyRequests через errors.Is.
func (e *RetryAfterError) Unwrap() error { return ErrTooManyRequests }

// GetAccrualOrder выполняет запрос к accrual системе для получения статуса заказа по его номеру.
// Возвращает ErrNotRegistered при HTTP 204 и RetryAfterError при HTTP 429.
func GetAccrualOrder(client *resty.Client, number string) (*AccrualResponse, error) {
	url := fmt.Sprintf("/api/orders/%s", number)
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, fmt.Errorf("accrual request: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var result AccrualResponse
		if err = json.Unmarshal(resp.Body(), &result); err != nil {
			return nil, fmt.Errorf("accrual decode: %w", err)
		}
		return &result, nil
	case http.StatusNoContent:
		return nil, ErrNotRegistered
	case http.StatusTooManyRequests:
		d := parseRetryAfter(resp.Header().Get("Retry-After"))
		return nil, &RetryAfterError{RetryAfter: d}
	default:
		return nil, fmt.Errorf("accrual unexpected status %d", resp.StatusCode())
	}
}

func parseRetryAfter(header string) time.Duration {
	if secs, err := strconv.Atoi(header); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return 60 * time.Second
}
