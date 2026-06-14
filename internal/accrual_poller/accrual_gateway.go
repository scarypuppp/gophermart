package accrual_poller

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

func GetAccrualOrder(client *resty.Client, number string) (*AccrualResponse, error) {
	request := client.R()
	url := fmt.Sprintf("/api/orders/%s", number)
	resp, err := request.Post(url)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode(), resp.String())
	}
	var result AccrualResponse
	if err = json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	return &result, err
}
