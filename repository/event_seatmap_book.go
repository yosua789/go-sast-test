package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/domain"
	"assist-tix/helper"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type EventSeatmapBookRepository interface {
	CreateSeatBook(ctx context.Context, tx pgx.Tx, eventId, venueSectorId string, reqs []domain.SeatmapParam) (err error)
	FindSeatStatusByRowColumnEventSectorId(ctx context.Context, tx pgx.Tx, eventId, venueSectorId string, seatRow, seatColumn int) (res model.EventSeatmapBook, err error)
	FindSeatBooksByEventSectorId(ctx context.Context, tx pgx.Tx, eventId, venueSectorId string) (seatmap map[string]model.EventSeatmapBook, err error)
}

type EventSeatmapBookRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewEventSeatmapBookRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) EventSeatmapBookRepository {
	return &EventSeatmapBookRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *EventSeatmapBookRepositoryImpl) CreateSeatBook(ctx context.Context, tx pgx.Tx, eventId, venueSectorId string, reqs []domain.SeatmapParam) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO event_seatmap_books (
		event_id,
		venue_sector_id,
		seat_row,
		seat_column,
		created_at
	) VALUES `

	var args []interface{}
	var placeholders []string

	for i, req := range reqs {
		base := i * 4
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, NOW())",
			base+1, base+2, base+3, base+4))

		args = append(args,
			eventId,
			venueSectorId,
			req.SeatRow,
			req.SeatColumn,
		)
	}

	query += strings.Join(placeholders, ",")

	if tx != nil {
		_, err = tx.Exec(ctx, query, args...)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, args...)
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				err = &lib.ErrorSeatIsAlreadyBooked
			}
		}
		return
	}

	return
}

func (r *EventSeatmapBookRepositoryImpl) FindSeatStatusByRowColumnEventSectorId(ctx context.Context, tx pgx.Tx, eventId, venueSectorId string, seatRow, seatColumn int) (res model.EventSeatmapBook, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT 
		id, 
		event_id,
		venue_sector_id,
		seat_row,
		seat_column,
		created_at
	FROM event_seatmap_books 
	WHERE event_id = $1 
	AND venue_sector_id = $2
	AND seat_row = $3 
	AND seat_column = $4`

	if tx != nil {
		err = tx.QueryRow(ctx, query, eventId, venueSectorId, seatRow, seatColumn).Scan(
			&res.ID,
			&res.EventID,
			&res.VenueSectorID,
			&res.SeatRow,
			&res.SeatColumn,
			&res.CreatedAt,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, eventId, venueSectorId, seatRow, seatColumn).Scan(
			&res.ID,
			&res.EventID,
			&res.VenueSectorID,
			&res.SeatRow,
			&res.SeatColumn,
			&res.CreatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = &lib.ErrorBookedSeatNotFound
		}
		return
	}

	return
}

func (r *EventSeatmapBookRepositoryImpl) FindSeatBooksByEventSectorId(ctx context.Context, tx pgx.Tx, eventId, venueSectorId string) (seatmap map[string]model.EventSeatmapBook, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	seatmap = make(map[string]model.EventSeatmapBook)

	query := `SELECT 
		id, 
		event_id,
		venue_sector_id,
		seat_row,
		seat_column,
		created_at
	FROM event_seatmap_books 
	WHERE event_id = $1 
	AND venue_sector_id = $2`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, eventId, venueSectorId)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, eventId, venueSectorId)
	}

	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var seatBook model.EventSeatmapBook
		rows.Scan(
			&seatBook.ID,
			&seatBook.EventID,
			&seatBook.VenueSectorID,
			&seatBook.SeatRow,
			&seatBook.SeatColumn,
			&seatBook.CreatedAt,
		)

		seatmap[helper.ConvertRowColumnKey(seatBook.SeatRow, seatBook.SeatColumn)] = seatBook
	}

	return
}
