package model

import (
	"time"
)

type ValueRecord struct {
	UUID      string    `json:"uuid"`
	Timestamp time.Time `json:"timestamp"`
	Value     []int64   `json:"value"`
}
