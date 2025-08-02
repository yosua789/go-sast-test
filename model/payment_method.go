package model

import (
	"database/sql"
	"time"
)

type PaymentMethod struct {
	ID       int
	Logo     string
	Name     string
	IsActive bool

	IsPaused     bool
	PauseMessage string
	PausedAt     sql.NullTime

	PaymentType    string
	PaymentGroup   string
	PaymentCode    string
	PaymentChannel string

	CreatedAt     time.Time
	UpdatedAt     sql.NullTime
	AdditionalFee float64
	IsPercentage  bool
}
