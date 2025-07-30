package entity

import (
	"database/sql"
	"time"
)

type EventTransaction struct {
	ID string

	OrderNumber       string
	Status            string
	StatusInformation string

	PaymentMethod  PaymentMethod
	Event          Event
	TicketCategory TicketCategory
	VenueSector    VenueSector

	PaymentExpiredAt time.Time

	PaidAt *time.Time

	TotalPrice           int
	TaxPercentage        sql.NullFloat64
	TotalTax             sql.NullInt32
	AdminFeePercentage   sql.NullFloat64
	AdditionalFeeDetails string
	TotalAdminFee        sql.NullInt32
	GrandTotal           int

	Fullname string
	Email    string
	// PhoneNumber string

	IsCompliment bool

	CreatedAt time.Time
	UpdatedAt *time.Time

	PaymentAdditionalInfo string // Virtual Account Number
	ChannelTransactionID  string // For lookup purposes
}
