package model

import (
	"database/sql"
	"time"
)

type EventTransactionItem struct {
	ID            int
	TransactionID string

	Quantity int

	SeatRow    int
	SeatColumn int
	SeatLabel  sql.NullString

	GarudaID sql.NullString

	Fullname    sql.NullString
	Email       sql.NullString
	PhoneNumber sql.NullString

	AdditionalInformation sql.NullString
	TotalPrice            int

	CreatedAt time.Time
	UpdatedAt *time.Time
}
