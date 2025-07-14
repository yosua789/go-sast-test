package model

import "time"

type EventTransactionItem struct {
	ID                    int
	TransactionID         string
	TicketCategoryID      string
	Quantity              int
	SeatRow               int
	SeatColumn            int
	AdditionalInformation string
	TotalPrice            int

	CreatedAt time.Time
	UpdatedAt *time.Time
}
