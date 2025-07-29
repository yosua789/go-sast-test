package entity

import (
	"database/sql"
)

type GroupPaymentMethod struct {
	PaymentGroup string
	Payments     []PaymentMethod
}

type PaymentMethod struct {
	ID   int
	Logo string
	Name string

	IsPaused     bool
	PauseMessage string
	PausedAt     sql.NullTime

	PaymentType    string
	PaymentGroup   string
	PaymentCode    string
	PaymentChannel string
}
