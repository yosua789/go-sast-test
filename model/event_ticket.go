package model

import (
	"database/sql"
	"time"
)

type EventTicket struct {
	ID int

	TicketOwnerEmail       string
	TicketOwnerFullname    string
	TicketOwnerPhoneNumber sql.NullString
	TicketOwnerGarudaId    sql.NullString

	EventID          string
	TicketCategoryID string
	TransactionID    string

	TicketNumber string
	TicketCode   string

	EventTime    time.Time
	EventVenue   string
	EventCity    string
	EventCountry string

	SectorName string
	AreaCode   string
	Entrance   string

	SeatRow    int
	SeatColumn int
	SeatLabel  string

	IsCompliment bool

	AdditionalInformation sql.NullString
}
