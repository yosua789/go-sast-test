package dto

type VirtualAccountSnapRequest struct {
	PartnerServiceID      string           `json:"partnerServiceId"`                // M, String, 8 - Derivative of X-PARTNER-ID, similar to company code, 8 digit left padding space
	CustomerNo            string           `json:"customerNo"`                      // M, String, 20 - virtualAccountNo 00000000000000000000
	VirtualAccountNo      string           `json:"virtualAccountNo"`                // M, String, 28 - partnerServiceId + customerNo
	VirtualAccountName    string           `json:"virtualAccountName,omitempty"`    // O, String, 255 - Payer. example: Jokul Doe
	VirtualAccountEmail   string           `json:"virtualAccountEmail,omitempty"`   // O, String, 255 - Email
	VirtualAccountPhone   string           `json:"virtualAccountPhone,omitempty"`   // O, String, 30 - Mobile Phone Number, Format: 62xxxxxxxxxxxxx
	TrxID                 string           `json:"trxId"`                           // M, String, 64 - Merchant Transaction Number
	TotalAmount           Amount           `json:"totalAmount"`                     // M, Object - Transaction amount
	BillDetails           *[]BillDetail    `json:"billDetails,omitempty"`           // O, List (max 24) - Array with maximum 24 Objects (Temporary Unavailable)
	BillDescription       *BillDescription `json:"billDescription,omitempty"`       // O, Object - Bill Description
	BillAmount            *Amount          `json:"billAmount,omitempty"`            // O, Object - Bill Amount
	AdditionalInfo        AdditionalInfo   `json:"additionalInfo"`                  // M, Object - Additional information
	FreeTexts             *[]FreeText      `json:"freeTexts,omitempty"`             // O, List (max 25) - Array with maximum 25 Objects
	VirtualAccountTrxType *string          `json:"virtualAccountTrxType,omitempty"` // O, String, 1 - Type of Virtual Account
	FeeAmount             *Amount          `json:"feeAmount,omitempty"`             // O, Object - Transaction Amount (Temporary Unavailable)
	ExpiredDate           string           `json:"expiredDate"`                     // M, String, 25 - Expiration date for Virtual Account. ISO-8601 format example: 2020-12-31T23:59:59-07:00
}

type Amount struct {
	Value    string `json:"value"`    // M, String, 16,2 - Amount with 2 decimal. Example: 10000.00 or for static VA: 0.00
	Currency string `json:"currency"` // M, String, 3 - Currency. Fixed value: IDR
}

type BillDetail struct {
	BillCode      *string `json:"billCode,omitempty"`      // O, String, 2 - Bill code for Customer choose
	BillNo        *string `json:"billNo,omitempty"`        // O, String, 18 - Bill number from Partner
	BillName      *string `json:"billName,omitempty"`      // O, String, 20 - Bill Name
	BillShortName *string `json:"billShortName,omitempty"` // O, String, 18 - Bill Name to be shown
}

type BillDescription struct {
	English        *string `json:"english,omitempty"`        // O, String, 18 - Bill Description in English
	Indonesia      *string `json:"indonesia,omitempty"`      // O, String, 18 - Bill Description in Bahasa
	BillSubCompany *string `json:"billSubCompany,omitempty"` // O, String, 5 - Sub company code
}

type FreeText struct {
	English   *string `json:"english,omitempty"`   // O, String, 32 - Will be shown in Channel (English)
	Indonesia *string `json:"indonesia,omitempty"` // O, String, 32 - Will be shown in Channel (Bahasa)
}

type AdditionalInfo struct {
	PaymentType string  `json:"paymentType"`       // M, String, 32 - PaymentType
	StoreID     *string `json:"storeId,omitempty"` // O, String, 32 - Only if merchant has branches and wish to create order using its branch ID
}

// paylabs VA

type PaylabsGenerateVAPayload struct {
	RequestID       string       `json:"requestId"`
	MerchantID      string       `json:"merchantId"`
	StoreID         string       `json:"storeId,omitempty"`
	PaymentType     string       `json:"paymentType"` // ex MandiriVA
	Amount          float64      `json:"amount"`      // ex 12000.00
	MerchantTradeNo string       `json:"merchantTradeNo"`
	NotifyURL       string       `json:"notifyUrl"`
	Payer           string       `json:"payer"`
	ProductName     string       `json:"productName"`
	ProductInfo     *ProductInfo `json:"productInfo,omitempty"` // Optional, for additional product details
}

type ProductInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Type     string  `json:"type"`               // ex: "ticket", "merchandise"
	Url      string  `json:"url,omitempty"`      // Optional, for merchandise or other products
	Quantity int     `json:"quantity,omitempty"` // Optional, for merchandise or other products
}

type PaylabsVACallbackRequest struct {
	RequestID       string      `json:"requestId"`
	ErrCode         string      `json:"errCode"`
	ErrCodeDes      string      `json:"errCodeDes"`
	MerchantID      string      `json:"merchantId"`
	StoreID         string      `json:"storeId"`
	PaymentType     string      `json:"paymentType"`
	Amount          float64     `json:"amount"`
	MerchantTradeNo string      `json:"merchantTradeNo"`
	PlatformTradeNo string      `json:"platformTradeNo"`
	CreateTime      string      `json:"createTime"`
	SuccessTime     string      `json:"successTime"`
	ProductName     string      `json:"productName"`
	ProductInfo     ProductInfo `json:"productInfo,omitempty"` // Optional, for additional product details
	TransFeeRate    float64     `json:"transFeeRate"`
	TransFeeAmount  float64     `json:"transFeeAmount"`
	TotalTransFee   float64     `json:"totalTransFee"`
	VatFee          float64     `json:"vatFee"`
}
type SnapCallbackAmount struct {
	Value    string `json:"value"`    // Paid amount (string, ISO4217, 16,2)
	Currency string `json:"currency"` // Fixed value: IDR
}

type SnapCallbackBillDescription struct {
	English   string `json:"english,omitempty"`   // * From Inquiry Response
	Indonesia string `json:"indonesia,omitempty"` // * From Inquiry Response
}

type SnapCallbackBillDetail struct {
	BillCode        string                       `json:"billCode,omitempty"`        // * From Inquiry Response
	BillNo          string                       `json:"billNo,omitempty"`          // * From Inquiry Response
	BillName        string                       `json:"billName,omitempty"`        // * From Inquiry Response
	BillShortName   string                       `json:"billShortName,omitempty"`   // * From Inquiry Response
	BillDescription *SnapCallbackBillDescription `json:"billDescription,omitempty"` // * From Inquiry Response
	BillSubCompany  string                       `json:"billSubCompany,omitempty"`  // * From Inquiry Response
	BillAmount      SnapCallbackAmount           `json:"billAmount"`                // Required
	AdditionalInfo  map[string]interface{}       `json:"additionalInfo"`            // * Unlimited additional info
	BillReferenceNo *string                      `json:"billReferenceNo,omitempty"` // * PJP generated bill auth code
}

type SnapCallbackFreeText struct {
	English   string `json:"english,omitempty"`   // * Up to 32 chars
	Indonesia string `json:"indonesia,omitempty"` // * Up to 32 chars
}

type SnapCallbackAdditionalInfo struct {
	TransFeeRate   *string `json:"transFeeRate,omitempty"`   // * Decimal(6,6)
	TransFeeAmount *string `json:"transFeeAmount,omitempty"` // * Decimal(12,2)
	TotalTransFee  *string `json:"totalTransFee,omitempty"`  // * Decimal(12,2)
	VatFee         *string `json:"vatFee,omitempty"`         // * Decimal(12,2)
	StoreID        *string `json:"storeId,omitempty"`        // * Only if merchant uses branch ID
	Channel        *string `json:"channel,omitempty"`        // * e.g., mobilephone
	DeviceID       *string `json:"deviceId,omitempty"`       // * e.g., 1234567890
	PaymentType    *string `json:"paymentType,omitempty"`    // * e.g., BCAVA
}

type SnapCallbackPaymentRequest struct {
	PartnerServiceId        string                      `json:"partnerServiceId"`                  // M: 8-char string, padded
	CustomerNo              string                      `json:"customerNo"`                        // M: 20-char, usually 00000000000000000000
	VirtualAccountNo        string                      `json:"virtualAccountNo"`                  // M: 28-char
	VirtualAccountName      *string                     `json:"virtualAccountName,omitempty"`      // * Max 255
	VirtualAccountEmail     *string                     `json:"virtualAccountEmail,omitempty"`     // * Max 255
	VirtualAccountPhone     *string                     `json:"virtualAccountPhone,omitempty"`     // * Format: 62xxxxxxxxxx
	TrxId                   *string                     `json:"trxId,omitempty"`                   // * C: Optional partner-generated trx ID
	PaymentRequestId        string                      `json:"paymentRequestId"`                  // M: 128-char max, required
	ChannelCode             *int                        `json:"channelCode,omitempty"`             // * O: ISO 18245 code
	HashedSourceAccountNo   *string                     `json:"hashedSourceAccountNo,omitempty"`   // * Max 32
	SourceBankCode          *string                     `json:"sourceBankCode,omitempty"`          // * 3-digit code
	PaidAmount              SnapCallbackAmount          `json:"paidAmount"`                        // M: Required
	CumulativePaymentAmount *SnapCallbackAmount         `json:"cumulativePaymentAmount,omitempty"` // * Optional
	PaidBills               *string                     `json:"paidBills,omitempty"`               // * Hex string
	TotalAmount             *SnapCallbackAmount         `json:"totalAmount,omitempty"`             // * Optional
	TrxDateTime             *string                     `json:"trxDateTime,omitempty"`             // * ISO-8601
	ReferenceNo             *string                     `json:"referenceNo,omitempty"`             // * Platform Order No
	JournalNum              *string                     `json:"journalNum,omitempty"`              // * Max 6-digit string
	PaymentType             *string                     `json:"paymentType,omitempty"`             // * Max 1-char string
	FlagAdvise              *string                     `json:"flagAdvise,omitempty"`              // * Y/N
	SubCompany              *string                     `json:"subCompany,omitempty"`              // * Max 5-char string
	BillDetails             []SnapCallbackBillDetail    `json:"billDetails,omitempty"`             // * Max 24 items
	FreeTexts               []SnapCallbackFreeText      `json:"freeTexts,omitempty"`               // * Max 25 items
	AdditionalInfo          *SnapCallbackAdditionalInfo `json:"additionalInfo"`                    // * Nested object
}

type PaylabsQRISRequest struct {
	MerchantID      string `json:"merchantId"`
	MerchantTradeNo string `json:"merchantTradeNo"`
	RequestID       string `json:"requestId"`
	PaymentType     string `json:"paymentType"`
	Amount          string `json:"amount"`
	ProductName     string `json:"productName"`
	Expire          int    `json:"expire"` // in second
	NotifyURL       string `json:"notifyUrl"`
}
