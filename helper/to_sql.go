package helper

import (
	"database/sql"
	"time"
)

func ToSQLString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{String: "", Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func ToSQLInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Int64: 0, Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

func ToSQLInt32(i int32) sql.NullInt32 {
	if i == 0 {
		return sql.NullInt32{Int32: 0, Valid: false}
	}
	return sql.NullInt32{Int32: i, Valid: true}
}

func ToSQLInt16(i int16) sql.NullInt16 {
	if i == 0 {
		return sql.NullInt16{Int16: 0, Valid: false}
	}
	return sql.NullInt16{Int16: i, Valid: true}
}

func ToSQLFloat64(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{Float64: 0, Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func ToSQLBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Bool: false, Valid: false}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

func ToSQLTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Time: time.Time{}, Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func ConvertNullTimeToPointer(nt sql.NullTime) *time.Time {
	if nt.Valid {
		truncatedTime := nt.Time.Truncate(time.Second)
		return &truncatedTime
	}
	return nil
}
