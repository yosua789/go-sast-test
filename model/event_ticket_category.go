package model

import (
	"database/sql"
	"time"
)

type EventTicketCategory struct {
	ID            string
	EventID       string
	VenueSectorId string
	Name          string
	Description   string
	Price         int

	TotalStock           int
	TotalPublicStock     int
	PublicStock          int
	TotalComplimentStock int
	ComplimentStock      int

	Code     string
	Entrance string

	CreatedAt time.Time
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}
