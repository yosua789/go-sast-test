package model

import (
	"database/sql"
	"time"
)

type Organizer struct {
	ID        string
	Name      string
	Slug      string
	Logo      string
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}
