package model

import "time"

type PaymentLog struct {
	ID            string    `json:"id"`
	Header        string    `json:"header"`
	Body          string    `json:"body"`
	Response      string    `json:"response"`
	CreatedAt     time.Time `json:"created_at"`
	ErrorResponse string    `json:"error_response"`
	Path          string    `json:"path"`
	ErrorCode     string    `json:"error_code"`
}
