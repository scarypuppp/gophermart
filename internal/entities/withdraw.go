package entities

import "time"

type Withdraw struct {
	order       string
	sum         int
	processedAt time.Time
}
