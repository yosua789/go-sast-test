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
	Status      string

	AdditionalInformation string

	IsActive bool

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
