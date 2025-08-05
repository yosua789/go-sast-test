package entity

import "database/sql"

type Sector struct {
	ID         string
	Name       string
	HasSeatmap bool
	Color      sql.NullString
	AreaCode   sql.NullString
}
