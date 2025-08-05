package async_order

type AdditionalFee struct {
	Name         string  `json:"name"`
	IsPercentage bool    `json:"is_percentage"`
	IsTax        bool    `json:"is_tax"`
	Value        float64 `json:"value"`
}
