package model

import (
	"database/sql"
	"time"
)

type Event struct {
	ID          string
	OrganizerID string
	Name        string
	Description string
	Banner      string
	EventTime   time.Time
	// Status      string
	VenueID string

	PublishStatus string
	IsSaleActive  bool

	AdditionalInformation string

	StartSaleAt sql.NullTime
	EndSaleAt   sql.NullTime

	CreatedAt time.Time
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}
