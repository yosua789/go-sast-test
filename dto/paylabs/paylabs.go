package paylabs

type VirtualAccountRequestSnap struct {
	PartnerServiceID      string             `json:"partnerServiceId"`
	CustomerNo            string             `json:"customerNo"`
	VirtualAccountNo      string             `json:"virtualAccountNo"`
	VirtualAccountName    string             `json:"virtualAccountName"`
	VirtualAccountEmail   string             `json:"virtualAccountEmail"`
	VirtualAccountPhone   string             `json:"virtualAccountPhone"`
	TrxID                 string             `json:"trxId"`
	TotalAmount           AmountSnap         `json:"totalAmount"`
	BillDetails           []BillDetailSnap   `json:"billDetails"`
	FreeTexts             []FreeTextSnap     `json:"freeTexts"`
	VirtualAccountTrxType string             `json:"virtualAccountTrxType"`
	FeeAmount             AmountSnap         `json:"feeAmount"`
	ExpiredDate           string             `json:"expiredDate"`
	AdditionalInfo        AdditionalInfoSnap `json:"additionalInfo"`
}

type AmountSnap struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type BillDetailSnap struct {
	BillCode        string                 `json:"billCode"`
	BillNo          string                 `json:"billNo"`
	BillName        string                 `json:"billName"`
	BillShortName   string                 `json:"billShortName"`
	BillDescription BillDescriptionSnap    `json:"billDescription"`
	BillSubCompany  string                 `json:"billSubCompany"`
	BillAmount      AmountSnap             `json:"billAmount"`
	AdditionalInfo  map[string]interface{} `json:"additionalInfo"` // Empty object, can hold dynamic keys
}

type BillDescriptionSnap struct {
	English   string `json:"english"`
	Indonesia string `json:"indonesia"`
}

type FreeTextSnap struct {
	English   string `json:"english"`
	Indonesia string `json:"indonesia"`
}

type AdditionalInfoSnap struct {
	PaymentType string `json:"paymentType"`
}
