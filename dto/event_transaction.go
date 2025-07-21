package dto

import "time"

type CreateEventTransaction struct {
	FullName    string `json:"fullname" validate:"required,alphaunicodespaces,max=255"`
	Email       string `json:"email" validate:"required,custom_email,max=255"`
	PhoneNumber string `json:"phone_number" validate:"required,custom_phone_number,max=50"`

	Items []OrderItemEventTransaction `json:"items" validate:"required"`

	PaymentMethod string `json:"payment_method" validate:"required"`
}

type OrderItemEventTransaction struct {
	SeatRow               int    `json:"seat_row" validate:"required"`
	SeatColumn            int    `json:"seat_column" validate:"required"`
	AdditionalInformation string `json:"additional_information" validate:"required"`
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
