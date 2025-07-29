package entity

import "database/sql"

type VenueSector struct {
	ID           string
	Venue        Venue
	Name         string
	SectorRow    sql.NullInt16
	SectorColumn sql.NullInt16
	Capacity     sql.NullInt32
	HasSeatmap   bool
	SectorColor  sql.NullString

	AreaCode sql.NullString
}
