package model

import (
	"database/sql"
	"time"
)

type Venue struct {
	ID        string
	VenueType string
	Name      string
	Country   string
	City      string
	Status    string
	Capacity  int

	CreatedAt time.Time
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}
