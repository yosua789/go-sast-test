package async_order

import "time"

type TransactionInvoice struct {
	TransactionID     string                       `json:"transaction_id"`
	OrderNumber       string                       `json:"order_number"`
	Status            string                       `json:"status"`
	AdditionalFees    []AdditionalFee              `json:"additional_fees"`
	Payment           PaymentInformation           `json:"payment"`
	DetailInformation DetailInformationTransaction `json:"detail_information"`
	Event             EventInformation             `json:"event"`
	ItemCount         int                          `json:"item_count"`
	ExpiredAt         time.Time                    `json:"expired_at"`
	CreatedAt         time.Time                    `json:"created_at"`
}
