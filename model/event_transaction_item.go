package model

import (
	"database/sql"
	"time"
)

type EventTransactionItem struct {
	ID                    int
	TransactionID         string
	TicketCategoryID      string
	Quantity              int
	SeatRow               int
	SeatColumn            int
	AdditionalInformation sql.NullString
	TotalPrice            int

	CreatedAt time.Time
	UpdatedAt *time.Time
}
