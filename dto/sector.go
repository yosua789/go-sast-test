package dto

type TicketCategorySectorResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	AreaCode string `json:"area_code"`
}

type SectorSeatmapResponse struct {
	Column int    `json:"column"`
	Label  string `json:"label"`
	Status string `json:"status"`
}

type SectorSeatmapRowResponse struct {
	Row   int                     `json:"row"`
	Seats []SectorSeatmapResponse `json:"seats"`
}

type EventSectorSeatmapResponse struct {
	ID       string                     `json:"id"`
	Name     string                     `json:"name"`
	Color    string                     `json:"color"`
	AreaCode string                     `json:"area_code"`
	Seatmap  []SectorSeatmapRowResponse `json:"seatmap"`
}
