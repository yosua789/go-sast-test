package entity

import (
	"time"
)

type OrderDetails struct {
	EventName             string    `json:"event_name"`              // event transaction -> event -> name
	VenueName             string    `json:"venue_name"`              // event transaction -> event -> venue
	EventTime             time.Time `json:"event_time"`              // event transaction -> event -> event_time
	TransactionDeadline   time.Time `json:"transaction_deadline"`    // event_transactions.payment_expired_at
	TransactionStatus     string    `json:"transaction_status"`      // event_transaction -> transaction -> transaction status
	PaymentMethod         string    `json:"payment_method"`          // if VA then return VA Number if qris return qr code string
	PaymentAdditionalInfo string    `json:"payment_additional_info"` // e.g. VA Number, QR Code
	GrandTotal            int       `json:"grand_total"`             // event_transaction -> transaction -> grand total
	TotalAdminFee         int       `json:"total_admin_fee"`         // event_transaction -> transaction -> total admin fee
	TotalTax              int       `json:"total_tax"`               // event_transaction -> transaction -> total tax
	TotalPrice            int       `json:"total_price"`             // event_transaction -> transaction -> total price
	TransactionQuantity   int       `json:"transaction_quantity"`    // event_transaction -> transaction -> item count
} // start from event_transaction
