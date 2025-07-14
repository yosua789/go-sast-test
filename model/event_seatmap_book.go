package model

import "time"

type EventSeatmapBook struct {
	ID            int
	EventID       string
	VenueSectorID string
	// EventTransactionID string

	SeatRow    int
	SeatColumn int

	CreatedAt time.Time
}
