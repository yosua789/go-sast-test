package model

import "time"

type EventTransactionGarudaID struct {
	ID        string    `json:"id" `
	EventID   string    `json:"event_id" `
	GarudaID  string    `json:"garuda_id" `
	CreatedAt time.Time `json:"created_at" `
}
