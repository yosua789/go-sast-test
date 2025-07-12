package model

import "time"

type VenueSector struct {
	ID           string
	VenueID      string
	Name         string
	SectorRow    int
	SectorColumn int
	Capacity     int
	IsActive     bool
	HasSeatmap   bool
	SectorColor  string
	AreaCode     string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
