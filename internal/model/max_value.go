package model

import "time"

type MaxValueRecord struct {
	UUID      string
	Timestamp time.Time
	MaxValue  int64
}
