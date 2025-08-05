package entity

import (
	"database/sql"
	"time"
)

type OrderDetails struct {
	EventName             string         `json:"event_name"` // event transaction -> event -> name
	VenueName             string         `json:"venue_name"` // event transaction -> event -> venue
	EventTime             time.Time      `json:"event_time"` // event transaction -> event -> event_time
	OrderNumber           string         `json:"order_number"`
	TransactionDeadline   time.Time      `json:"transaction_deadline"`    // event_transactions.payment_expired_at
	TransactionStatus     string         `json:"transaction_status"`      // event_transaction -> transaction -> transaction status
	TicketCategoryName    string         `json:"ticket_category_name"`    // event_transaction -> transaction -> transaction status
	PaymentMethod         string         `json:"payment_method"`          // if VA then return VA Number if qris return qr code string
	PaymentAdditionalInfo sql.NullString `json:"payment_additional_info"` // e.g. VA Number, QR Code
	PaymentPaidAt         *time.Time     `json:"payment_paid_at"`
	GrandTotal            int            `json:"grand_total"`          // event_transaction -> transaction -> grand total
	TotalAdminFee         int            `json:"total_admin_fee"`      // event_transaction -> transaction -> total admin fee
	TotalTax              int            `json:"total_tax"`            // event_transaction -> transaction -> total tax
	TotalPrice            int            `json:"total_price"`          // event_transaction -> transaction -> total price
	TransactionQuantity   int            `json:"transaction_quantity"` // event_transaction -> transaction -> item count
	Country               string         `json:"country"`              // event transaction -> user -> country
	City                  string         `json:"city"`                 // event transaction -> user -> city
	PGAdditionalFee       int            `json:"pg_additional_fee"`    // event transaction -> transaction -> additional fee for payment gateway
} // start from event_transaction

type AdditionalPaymentInfo struct {
	Name            string  `json:"name"`
	IsTax           bool    `json:"is_tax"`
	IsPercentage    bool    `json:"is_percentage"`
	Value           float64 `json:"value"`
	CalculatedValue float64 `json:"calculated_value"` // calculated value based on total price
}
