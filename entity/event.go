package entity

import (
	"database/sql"
	"time"
)

type Event struct {
	ID string

	Organizer Organizer
	Venue     Venue

	Name        string
	Description string
	Banner      string
	EventTime   time.Time

	PublishStatus string
	IsSaleActive  bool

	AdditionalInformation string

	StartSaleAt sql.NullTime
	EndSaleAt   sql.NullTime

	CreatedAt time.Time
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}

type PaginatedEvents struct {
	Events     []Event
	Pagination Pagination
}

type EventVenueSector struct {
	ID           int
	SeatRow      int
	SeatColumn   int
	SeatRowLabel int
	Label        string
	Status       string
}
