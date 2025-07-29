package dto

import "time"

type CreateEventTransaction struct {
	Fullname string `json:"fullname" validate:"required,alphaunicodespaces,max=255"`
	Email    string `json:"email" validate:"required,custom_email,max=255"`
	// PhoneNumber string `json:"phone_number" validate:"required,custom_phone_number,max=50"`

	Items []OrderItemEventTransaction `json:"items" validate:"required,dive"`

	PaymentMethod string `json:"payment_method" validate:"required"`
}

type OrderItemEventTransaction struct {
	SeatRow    int `json:"seat_row" validate:"omitempty,min=1"`
	SeatColumn int `json:"seat_column" validate:"omitempty,min=1"`

	FullName    string `json:"fullname" validate:"omitempty,alphaunicodespaces,max=255"`
	Email       string `json:"email" validate:"omitempty,custom_email,max=255"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,custom_phone_number,max=50"`

	GarudaID              string `json:"garuda_id" validate:"omitempty,max=20"`
	AdditionalInformation string `json:"additional_information" validate:""`
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
	AccessToken        string  `json:"access_token"`

	ExpiredAt time.Time `json:"payment_expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

type GetTransactionDetails struct {
	TransactionID string `uri:"transactionId" binding:"required,min=1,uuid"`
}
