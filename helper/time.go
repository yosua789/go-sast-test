package helper

import (
	"database/sql"
	"time"
)

func FromNilTime(
	data sql.NullTime,
) *time.Time {
	if data.Valid {
		return &data.Time
	}
	return nil
}
