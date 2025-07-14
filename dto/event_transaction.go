package dto

import "time"

type CreateEventTransaction struct {
	FullName    string
	Email       string
	PhoneNumber string

	Items []OrderItemEventTransaction

	PaymentMethod string
}

type OrderItemEventTransaction struct {
	SeatRow               int
	SeatColumn            int
	AdditionalInformation string
}

type EventTransactionResponse struct {
	InvoiceNumber      string  `json:"invoice_number"`
	PaymentMethod      string  `json:"payment_method"`
	TotalPrice         int     `json:"total_price"`
	TaxPercentage      float32 `json:"tax_percentage"`
	TotalTax           int     `json:"total_tax"`
	AdminFeePercentage float32 `json:"admin_fee_percentage"`
	TotalAdminFee      int     `json:"total_admin_fee"`
	GrandTotal         int     `json:"grand_total"`

	ExpiredAt time.Time `json:"payment_expired_at"`
	CreatedAt time.Time `json:"created_at"`
}
