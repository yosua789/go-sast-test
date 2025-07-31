package event

import "time"

type TransactionETicket struct {
	TransactionID     string                       `json:"transaction_id"`
	TicketNumber      string                       `json:"ticket_number"`
	TicketCode        string                       `json:"ticket_code"`
	Payment           PaymentInformation           `json:"payment"`
	DetailInformation DetailInformationTransaction `json:"detail_information"`
	Event             EventInformation             `json:"event"`
	CreatedAt         time.Time                    `json:"created_at"`
}
