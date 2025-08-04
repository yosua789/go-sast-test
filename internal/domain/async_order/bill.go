package async_order

import "time"

// Order mean, transaction status `PENDING`
type TransactionBill struct {
	TransactionID          string                       `json:"transaction_id"`
	OrderNumber            string                       `json:"order_number"`
	Status                 string                       `json:"status"`
	Payment                PaymentInformation           `json:"payment"`
	DetailInformation      DetailInformationTransaction `json:"detail_information"`
	Event                  EventInformation             `json:"event"`
	TransactionAccessToken string                       `json:"transaction_access_token"`
	ItemCount              int                          `json:"item_count"`
	ExpiredAt              time.Time                    `json:"expired_at"`
	CreatedAt              time.Time                    `json:"created_at"`
}

type EventInformation struct {
	ID             string    `json:"id"`
	BannerFilename string    `json:"banner_filename"`
	Name           string    `json:"name"`
	Time           time.Time `json:"time"`
}

type DetailInformationTransaction struct {
	BookEmail       string `json:"book_email"`
	BookName        string `json:"book_name"`
	BookPhoneNumber string `json:"book_phone_number,omitempty"`
	BookGarudaID    string `json:"book_garuda_id,omitempty"`
	UseGarudaId     bool   `json:"use_garuda_id"`

	TicketCategory TicketCategoryInformation `json:"ticket_category"`
	Location       LocationInformation       `json:"location"`
}

type LocationInformation struct {
	VenueType string `json:"venue_type"`
	VenueName string `json:"venue_name"`
	Country   string `json:"country"`
	City      string `json:"city"`
}

type TicketCategoryInformation struct {
	Code     string       `json:"code"`
	Price    int          `json:"price"`
	Name     string       `json:"name"`
	Entrance string       `json:"entrance"`
	Sector   TicketSector `json:"sector"`
}

type TicketSector struct {
	Name string `json:"name"`
}

type PaymentInformation struct {
	// Method      string `json:"method"`
	Type    string `json:"type"`
	Group   string `json:"group"`
	Code    string `json:"code"`
	Channel string `json:"channel"`

	DisplayName string `json:"display_name"`

	PaymentAdditionalInformation string `json:"payment_additional_information"`

	GrandTotal int `json:"grand_total"`
}
