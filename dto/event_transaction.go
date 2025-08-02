package dto

import (
	"assist-tix/entity"
	"time"
)

type CreateEventTransaction struct {
	Fullname string `json:"fullname" validate:"required,alphaunicodespaces,min=3,max=255"`
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
	TransactionID      string  `json:"transaction_id"`
	OrderNumber        string  `json:"order_number"`
	PaymentMethod      string  `json:"payment_method"`
	TotalPrice         int     `json:"total_price"`
	TaxPercentage      float32 `json:"tax_percentage"`
	TotalTax           int     `json:"total_tax"`
	AdminFeePercentage float32 `json:"admin_fee_percentage"`
	TotalAdminFee      int     `json:"total_admin_fee"`
	GrandTotal         int     `json:"grand_total"`
	AccessToken        string  `json:"access_token"`
	PgAdditionalFee    int     `json:"pg_additional_fee"`

	ExpiredAt time.Time `json:"payment_expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

type EventGrouppedPaymentMethodsResponse struct {
	PaymentGroup string                       `json:"payment_group"`
	Payments     []EventPaymentMethodResponse `json:"payments"`
}

type EventPaymentMethodResponse struct {
	Name string `json:"name"`
	Logo string `json:"logo"`

	IsPaused     bool       `json:"is_paused"`
	PauseMessage string     `json:"pause_message"`
	PausedAt     *time.Time `json:"paused_at"`

	PaymentType   string  `json:"payment_type"`
	PaymentCode   string  `json:"payment_code"`
	AdditionalFee float64 `json:"additional_fee"`
	IsPercentage  bool    `json:"is_percentage"`
}
type GetTransactionDetails struct {
	TransactionID string `uri:"transactionId" binding:"required,min=1,uuid"`
}

type OrderDetails struct {
	EventName             string                         `json:"event_name"`              // event transaction -> event -> name
	VenueName             string                         `json:"venue_name"`              // event transaction -> event -> venue
	EventTime             time.Time                      `json:"event_time"`              // event transaction -> event -> event_time
	TransactionDeadline   time.Time                      `json:"transaction_deadline"`    // event_transactions.payment_expired_at
	TransactionStatus     string                         `json:"transaction_status"`      // event_transaction -> transaction -> transaction status
	PaymentMethod         string                         `json:"payment_method"`          // if VA then return VA Number if qris return qr code string
	PaymentAdditionalInfo string                         `json:"payment_additional_info"` // e.g. VA Number, QR Code
	GrandTotal            int                            `json:"grand_total"`             // event_transaction -> transaction -> grand total
	TotalAdminFee         int                            `json:"total_admin_fee"`         // event_transaction -> transaction -> total admin fee
	TotalTax              int                            `json:"total_tax"`               // event_transaction -> transaction -> total tax
	TotalPrice            int                            `json:"total_price"`             // event_transaction -> transaction -> total price
	TransactionQuantity   int                            `json:"transaction_quantity"`    // event_transaction -> transaction -> item count
	Country               string                         `json:"country"`                 // event transaction -> user -> country
	City                  string                         `json:"city"`                    // event transaction -> user -> city
	AdditionalPayment     []entity.AdditionalPaymentInfo `json:"additional_payment"`      // event transaction -> transaction -> additional payment info
	PGAdditionalFee       int                            `json:"pg_additional_fee"`       // event transaction -> transaction -> additional fee for payment gateway
}
