package model

import "time"

type EventTransaction struct {
	ID                string
	InvoiceNumber     string
	Status            string
	StatusInformation string
	PaymentMethod     string
	PaymentChannel    string
	PaymentExpiredAt  time.Time
	PaidAt            *time.Time

	TotalPrice int
	// TaxPercentage      float32
	TotalTax int
	// AdminFeePercentage float32
	AdditionalFeeDetails string
	TotalAdminFee        int
	GrandTotal           int

	Fullname string
	Email    string
	// PhoneNumber string

	IsCompliment bool

	CreatedAt time.Time
	UpdatedAt *time.Time

	VANumber             string // Virtual Account Number
	ChannelTransactionID string // For lookup purposes
}
