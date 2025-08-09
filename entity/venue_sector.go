package entity

import "database/sql"

type VenueSector struct {
	ID           sql.NullString
	Venue        Venue
	Name         sql.NullString
	SectorRow    sql.NullInt16
	SectorColumn sql.NullInt16
	Capacity     sql.NullInt32
	HasSeatmap   sql.NullBool
	SectorColor  sql.NullString

	AreaCode sql.NullString
}
