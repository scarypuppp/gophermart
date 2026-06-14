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

var ErrNotRegistered = errors.New("order not registered in accrual")
var ErrTooManyRequests = errors.New("accrual rate limit exceeded")

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

type RetryAfterError struct {
	RetryAfter time.Duration
}

func (e *RetryAfterError) Error() string {
	return fmt.Sprintf("%s: retry after %s", ErrTooManyRequests, e.RetryAfter)
}

func (e *RetryAfterError) Unwrap() error { return ErrTooManyRequests }

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
