package async_order

import (
	"assist-tix/entity"
	"assist-tix/model"
)

type AsyncOrder struct {
	ClientIP               string                    `json:"client_ip"`
	OrderInformationBookID int                       `json:"order_information_book_id"`
	UseGarudaId            bool                      `json:"use_garuda_id"`
	ItemCount              int                       `json:"item_count"`
	TransactionAccessToken string                    `json:"transaction_access_token"`
	PaymentMethod          model.PaymentMethod       `json:"payment_method"`
	Event                  model.Event               `json:"event"`
	Transaction            model.EventTransaction    `json:"transaction"`
	TicketCategory         model.EventTicketCategory `json:"ticket_category"`
	VenueSector            entity.VenueSector        `json:"venue_sector"`
}
