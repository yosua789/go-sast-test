package event

import "time"

type TransactionETicket struct {
	TicketID      int    `json:"ticket_id"`
	TransactionID string `json:"transaction_id"`
	TicketNumber  string `json:"ticket_number"`
	TicketCode    string `json:"ticket_code"`

	TicketSeatRow    int    `json:"ticket_seat_row"`
	TicketSeatColumn int    `json:"ticket_seat_column"`
	TicketSeatLabel  string `json:"ticket_seat_label"`

	Payment           PaymentInformation           `json:"payment"`
	DetailInformation DetailInformationTransaction `json:"detail_information"`
	Event             EventInformation             `json:"event"`
	CreatedAt         time.Time                    `json:"created_at"`
}
