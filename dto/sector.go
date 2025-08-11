package dto

import "time"

type VenueSectorResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	IsActive     bool       `json:"is_active"`
	HasSeatmap   bool       `json:"has_seatmap"`
	SectorRow    int        `json:"sector_row"`
	SectorColumn int        `json:"sector_column"`
	Capacity     int        `json:"capacity"`
	SectorColor  string     `json:"sector_color"`
	AreaCode     string     `json:"area_code"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type TicketCategorySectorResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	AreaCode   string `json:"area_code"`
	HasSeatmap bool   `json:"has_seatmap"`
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

type SectorSeatmapRowColumnResponse struct {
	Row      int    `json:"row"`
	Column   int    `json:"column"`
	RowLabel int    `json:"row_label"`
	Label    string `json:"label"`
	Status   string `json:"status"`
}
