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
	HeldAt      sql.NullTime
	Status      string
	CreatedAt   time.Time
	UpdatedAt   sql.NullTime
	DeletedAt   sql.NullTime
}
