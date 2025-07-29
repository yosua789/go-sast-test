package model

import "time"

type ReleaseTransactionJob struct {
	TransactionID string    `json:"transaction_id"`
	CreatedAt     time.Time `json:"created_at"`
}
