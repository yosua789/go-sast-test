package model

import "time"

type EventTransactionGarudaID struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EventID   string    `json:"event_id" gorm:"type:uuid;not null"`
	GarudaID  string    `json:"garuda_id" gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;default:current_timestamp"`
}
