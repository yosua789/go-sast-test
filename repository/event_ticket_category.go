package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/domain"
	"assist-tix/entity"
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
	"github.com/rs/zerolog/log"
)

type EventTicketCategoryRepository interface {
	Create(ctx context.Context, tx pgx.Tx, ticketCategory model.EventTicketCategory) (err error)
	FindTicketSectorsByEventId(ctx context.Context, tx pgx.Tx, eventId string) (ticketCategories []entity.TicketCategory, err error)
	FindByEventId(ctx context.Context, tx pgx.Tx, eventId string) (ticketCategories []model.EventTicketCategory, err error)
	FindByIdAndEventId(ctx context.Context, tx pgx.Tx, eventId string, ticketCategoryId string) (res model.EventTicketCategory, err error)
	FindSeatmapByEventSectorId(ctx context.Context, tx pgx.Tx, eventId, sectorId string) (seats []entity.EventVenueSector, err error)
	FindSeatmapStatusByEventSectorId(ctx context.Context, tx pgx.Tx, eventId, sectorId string, reqs []domain.SeatmapParam) (seatmap map[string]entity.EventVenueSector, err error)
	BuyPublicTicketById(ctx context.Context, tx pgx.Tx, eventId, ticketCategoryId string, newStock int) (err error)
	SoftDelete(ctx context.Context, tx pgx.Tx, ticketCategoryId string) (err error)
	FindLowestPriceTicketByEventIds(ctx context.Context, tx pgx.Tx, eventIds ...string) (res map[string]int, err error)
	FindTotalSaleTicketByEventIds(ctx context.Context, tx pgx.Tx, eventIds ...string) (res map[string]int, err error)
	FindSeatByRowsColumnsEventSectorId(ctx context.Context, tx pgx.Tx, eventId, sectorId string, seatmaps ...domain.SeatmapParam) (seats map[string]entity.EventVenueSector, err error)
	FindNAvailableSeatAfterSectorRowColumn(ctx context.Context, tx pgx.Tx, eventId, sectorId string, seatCount, seatRow, seatColumn int) (seats []entity.EventVenueSector, err error)
}

type EventTicketCategoryRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewEventTicketCategoryRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) EventTicketCategoryRepository {
	return &EventTicketCategoryRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *EventTicketCategoryRepositoryImpl) Create(ctx context.Context, tx pgx.Tx, ticketCategory model.EventTicketCategory) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO event_ticket_categories (event_id, venue_sector_id, name, description, price, total_stock, total_public_stock, public_stock, total_compliment_stock, compliment_stock, code, entrance, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())`

	if tx != nil {
		_, err = tx.Exec(ctx, query,
			ticketCategory.EventID,
			ticketCategory.VenueSectorId,
			ticketCategory.Name,
			ticketCategory.Description,
			ticketCategory.Price,
			ticketCategory.TotalStock,
			ticketCategory.TotalPublicStock,
			ticketCategory.PublicStock,
			ticketCategory.TotalComplimentStock,
			ticketCategory.ComplimentStock,
			ticketCategory.Code,
			ticketCategory.Entrance,
		)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query,
			ticketCategory.EventID,
			ticketCategory.VenueSectorId,
			ticketCategory.Name,
			ticketCategory.Description,
			ticketCategory.Price,
			ticketCategory.TotalStock,
			ticketCategory.TotalPublicStock,
			ticketCategory.PublicStock,
			ticketCategory.TotalComplimentStock,
			ticketCategory.ComplimentStock,
			ticketCategory.Code,
			ticketCategory.Entrance,
		)
	}

	if err != nil {
		return err
	}

	return
}

func (r *EventTicketCategoryRepositoryImpl) FindTicketSectorsByEventId(ctx context.Context, tx pgx.Tx, eventId string) (ticketCategories []entity.TicketCategory, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT 
		etc.id,
		etc.name, 
		etc.description,
		etc.price, 
		etc.total_stock, 
		etc.total_public_stock, 
		etc.public_stock, 
		etc.total_compliment_stock, 
		etc.compliment_stock, 
		etc.code, 
		etc.entrance,

		vs.id as sector_id,
		vs.name as sector_name,
		vs.has_seatmap as sector_has_seatmap,
		vs.sector_color as sector_color,
		vs.area_code as sector_area_code

	FROM event_ticket_categories AS etc
	INNER JOIN venue_sectors AS vs
		ON etc.venue_sector_id = vs.id
	WHERE etc.event_id = $1 AND etc.deleted_at IS NULL`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, eventId)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, eventId)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var ticketCategory entity.TicketCategory
		rows.Scan(
			&ticketCategory.ID,
			&ticketCategory.Name,
			&ticketCategory.Description,
			&ticketCategory.Price,
			&ticketCategory.TotalStock,
			&ticketCategory.TotalPublicStock,
			&ticketCategory.PublicStock,
			&ticketCategory.TotalComplimentStock,
			&ticketCategory.ComplimentStock,
			&ticketCategory.Code,
			&ticketCategory.Entrance,

			&ticketCategory.Sector.ID,
			&ticketCategory.Sector.Name,
			&ticketCategory.Sector.HasSeatmap,
			&ticketCategory.Sector.Color,
			&ticketCategory.Sector.AreaCode,
		)

		ticketCategories = append(ticketCategories, ticketCategory)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("find ticket by event error")
		return ticketCategories, rows.Err()
	}

	return
}

func (r *EventTicketCategoryRepositoryImpl) FindByEventId(ctx context.Context, tx pgx.Tx, eventId string) (ticketCategories []model.EventTicketCategory, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT 
		id,
		name, 
		description,
		price, 
		total_stock, 
		total_public_stock, 
		public_stock, 
		total_compliment_stock, 
		compliment_stock, 
		code, 
		entrance, 
		created_at,
		updated_at
	FROM event_ticket_categories 
	WHERE event_id = $1 AND deleted_at IS NULL`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, eventId)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, eventId)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var ticketCategory model.EventTicketCategory
		rows.Scan(
			&ticketCategory.ID,
			&ticketCategory.Name,
			&ticketCategory.Description,
			&ticketCategory.Price,
			&ticketCategory.TotalStock,
			&ticketCategory.TotalPublicStock,
			&ticketCategory.PublicStock,
			&ticketCategory.TotalComplimentStock,
			&ticketCategory.ComplimentStock,
			&ticketCategory.Code,
			&ticketCategory.Entrance,
			&ticketCategory.CreatedAt,
			&ticketCategory.UpdatedAt,
		)

		ticketCategories = append(ticketCategories, ticketCategory)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("find ticket by event error")
		return ticketCategories, rows.Err()
	}

	return
}

func (r *EventTicketCategoryRepositoryImpl) FindByIdAndEventId(ctx context.Context, tx pgx.Tx, eventId string, ticketCategoryId string) (res model.EventTicketCategory, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT 
		id,
		venue_sector_id,
		name, 
		description,
		price, 
		total_stock, 
		total_public_stock, 
		public_stock, 
		total_compliment_stock, 
		compliment_stock, 
		code, 
		entrance, 
		created_at,
		updated_at
	FROM event_ticket_categories 
	WHERE event_id = $1 AND id = $2 AND deleted_at IS NULL`

	if tx != nil {
		err = tx.QueryRow(ctx, query, eventId, ticketCategoryId).Scan(
			&res.ID,
			&res.VenueSectorId,
			&res.Name,
			&res.Description,
			&res.Price,
			&res.TotalStock,
			&res.TotalPublicStock,
			&res.PublicStock,
			&res.TotalComplimentStock,
			&res.ComplimentStock,
			&res.Code,
			&res.Entrance,
			&res.CreatedAt,
			&res.UpdatedAt,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, eventId, ticketCategoryId).Scan(
			&res.ID,
			&res.VenueSectorId,
			&res.Name,
			&res.Description,
			&res.Price,
			&res.TotalStock,
			&res.TotalPublicStock,
			&res.PublicStock,
			&res.TotalComplimentStock,
			&res.ComplimentStock,
			&res.Code,
			&res.Entrance,
			&res.CreatedAt,
			&res.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return res, &lib.ErrorTicketCategoryNotFound
		}
		return res, err
	}

	return
}

func (r *EventTicketCategoryRepositoryImpl) SoftDelete(ctx context.Context, tx pgx.Tx, ticketCategoryId string) (err error) {

	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE event_ticket_categories SET
		deleted_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL`

	// var cmdTag pgCommandTag
	if tx != nil {
		_, err = tx.Exec(ctx, query, ticketCategoryId)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, ticketCategoryId)
	}

	return
}

func (r *EventTicketCategoryRepositoryImpl) FindSeatmapByEventSectorId(ctx context.Context, tx pgx.Tx, eventId, sectorId string) (seats []entity.EventVenueSector, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT 
		vssm.id, 
		vssm.seat_row, 
		vssm.seat_column, 
		CASE 
			WHEN vssm.label != evssm.label THEN evssm.label
			ELSE vssm.label
		END AS seat_final_label,
		CASE 
			WHEN vssm.status != evssm.status THEN 
				CASE 
					WHEN evssm.status IN ('PREBOOK', 'COMPLIMENT') THEN 'UNAVAILABLE'
					ELSE evssm.status
				END 
			ELSE vssm.status
		END AS seat_final_status
	FROM venue_sector_seatmap_matrix vssm 
	LEFT JOIN event_venue_sector_seatmap_matrix evssm 
		ON vssm.sector_id = evssm.sector_id 
		AND vssm.seat_row = evssm.seat_row 
		AND vssm.seat_column = evssm.seat_column
		AND evssm.event_id = $1
	WHERE vssm.sector_id = $2
	ORDER BY vssm.seat_row ASC, vssm.seat_column ASC`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, eventId, sectorId)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, eventId, sectorId)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var sectorSeatmap entity.EventVenueSector
		rows.Scan(
			&sectorSeatmap.ID,
			&sectorSeatmap.SeatRow,
			&sectorSeatmap.SeatColumn,
			&sectorSeatmap.Label,
			&sectorSeatmap.Status,
		)

		seats = append(seats, sectorSeatmap)
	}

	return
}

func (r *EventTicketCategoryRepositoryImpl) BuyPublicTicketById(ctx context.Context, tx pgx.Tx, eventId, ticketCategoryId string, buyTicket int) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE event_ticket_categories 
		SET public_stock = public_stock - $1, 
		updated_at = NOW() 
	WHERE event_id = $2 
		AND id = $3
		AND public_stock >= $1
		AND deleted_at IS NULL`

	var cmdTag pgconn.CommandTag
	if tx != nil {
		cmdTag, err = tx.Exec(ctx, query, buyTicket, eventId, ticketCategoryId)
	} else {
		cmdTag, err = r.WrapDB.Postgres.Exec(ctx, query, buyTicket, eventId, ticketCategoryId)
	}

	if err != nil {
		return
	}

	if cmdTag.RowsAffected() == 0 {
		err = &lib.ErrorTicketIsOutOfStock
		return
	}

	return
}

// Find seatmap by specific event, venue sector and seat row & colum
func (r *EventTicketCategoryRepositoryImpl) FindSeatmapStatusByEventSectorId(ctx context.Context, tx pgx.Tx, eventId, sectorId string, reqs []domain.SeatmapParam) (seatmap map[string]entity.EventVenueSector, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	seatmap = make(map[string]entity.EventVenueSector)

	query := `SELECT 
		vssm.id, 
		vssm.seat_row, 
		vssm.seat_column, 
		CASE 
			WHEN vssm.label != evssm.label THEN evssm.label
			ELSE vssm.label
		END AS seat_final_label,
		CASE 
			WHEN vssm.status != evssm.status THEN 
				CASE 
					WHEN evssm.status IN ('PREBOOK', 'COMPLIMENT') THEN 'UNAVAILABLE'
					ELSE evssm.status
				END 
			ELSE vssm.status
		END AS seat_final_status
	FROM venue_sector_seatmap_matrix vssm 
	LEFT JOIN event_venue_sector_seatmap_matrix evssm 
		ON vssm.sector_id = evssm.sector_id 
		AND vssm.seat_row = evssm.seat_row 
		AND vssm.seat_column = evssm.seat_column
		AND evssm.event_id = $1
	WHERE vssm.sector_id = $2
		AND (vssm.seat_row, vssm.seat_column) IN (`

	var args []interface{}
	var placeholders []string

	args = append(args, eventId, sectorId)

	for i, req := range reqs {
		base := (i * 2) + 2
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)",
			base+1, base+2))

		args = append(args,
			req.SeatRow,
			req.SeatColumn,
		)
	}

	query += strings.Join(placeholders, ",")
	query += `) ORDER BY vssm.seat_row ASC, vssm.seat_column ASC`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, args...)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var sectorSeatmap entity.EventVenueSector
		rows.Scan(
			&sectorSeatmap.ID,
			&sectorSeatmap.SeatRow,
			&sectorSeatmap.SeatColumn,
			&sectorSeatmap.Label,
			&sectorSeatmap.Status,
		)

		seatmap[helper.ConvertRowColumnKey(sectorSeatmap.SeatRow, sectorSeatmap.SeatColumn)] = sectorSeatmap
	}

	return
}

// Find lowest price ticket category by event Ids
// Returning map of ticketId
func (r *EventTicketCategoryRepositoryImpl) FindLowestPriceTicketByEventIds(ctx context.Context, tx pgx.Tx, eventIds ...string) (res map[string]int, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	res = make(map[string]int)

	if len(eventIds) <= 0 {
		return
	}

	query := `SELECT
		event_id, 
		min(price) as lowest_price
	FROM event_ticket_categories  
	WHERE event_id IN (`

	var args []interface{}
	var placeholders []string

	for i, val := range eventIds {
		base := i
		placeholders = append(placeholders, fmt.Sprintf("$%d",
			base+1))

		args = append(args, val)
	}

	query += strings.Join(placeholders, ",")
	query += `) GROUP BY event_id`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, args...)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var eventId string
		var lowestPrice int
		rows.Scan(
			&eventId,
			&lowestPrice,
		)

		res[eventId] = lowestPrice
	}

	return
}

func (r *EventTicketCategoryRepositoryImpl) FindTotalSaleTicketByEventIds(ctx context.Context, tx pgx.Tx, eventIds ...string) (res map[string]int, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	res = make(map[string]int)

	if len(eventIds) <= 0 {
		return
	}

	query := `SELECT 
		event_id, 
		SUM(total_public_stock) as total_public_sale 
	FROM event_ticket_categories
	WHERE deleted_at IS NULL 
	AND event_id IN (`

	var args []interface{}
	var placeholders []string

	for i, val := range eventIds {
		base := i
		placeholders = append(placeholders, fmt.Sprintf("$%d",
			base+1))

		args = append(args, val)
	}

	query += strings.Join(placeholders, ",")
	query += `) GROUP BY event_id`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, args...)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var eventId string
		var totalPublicSale int

		rows.Scan(
			&eventId,
			&totalPublicSale,
		)

		res[eventId] = totalPublicSale
	}

	return
}

// It doesn't check seat is booked or not
func (r *EventTicketCategoryRepositoryImpl) FindSeatByRowsColumnsEventSectorId(ctx context.Context, tx pgx.Tx, eventId, sectorId string, seatmaps ...domain.SeatmapParam) (seats map[string]entity.EventVenueSector, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	var conditions []string
	var args []interface{}
	argIndex := 1

	query := `SELECT 
		vssm.id, 
		vssm.seat_row, 
		vssm.seat_column, 
		CASE 
			WHEN vssm.label != evssm.label THEN evssm.label
			ELSE vssm.label
		END AS seat_final_label,
		CASE 
			WHEN vssm.status != evssm.status THEN 
				CASE 
					WHEN evssm.status IN ('PREBOOK', 'COMPLIMENT') THEN 'UNAVAILABLE'
					ELSE evssm.status
				END 
			ELSE vssm.status
		END AS seat_final_status
	FROM venue_sector_seatmap_matrix vssm 
	LEFT JOIN event_venue_sector_seatmap_matrix evssm 
		ON vssm.sector_id = evssm.sector_id 
		AND vssm.seat_row = evssm.seat_row 
		AND vssm.seat_column = evssm.seat_column
		AND evssm.event_id = $1
	WHERE vssm.sector_id = $2
		AND (vssm.seat_row, vssm.seat_column) IN (`

	args = append(args, eventId, sectorId)
	argIndex += 2

	for _, seat := range seatmaps {
		conditions = append(conditions, fmt.Sprintf(`($%d, $%d)`, argIndex, argIndex+1))
		args = append(args, seat.SeatRow, seat.SeatColumn)
		argIndex += 2
	}

	query += strings.Join(conditions, ",")
	query += `) ORDER BY vssm.seat_row ASC, vssm.seat_column ASC`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, args...)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var sectorSeatmap entity.EventVenueSector
		rows.Scan(
			&sectorSeatmap.ID,
			&sectorSeatmap.SeatRow,
			&sectorSeatmap.SeatColumn,
			&sectorSeatmap.Label,
			&sectorSeatmap.Status,
		)

		seats[helper.ConvertRowColumnKey(sectorSeatmap.SeatRow, sectorSeatmap.SeatColumn)] = sectorSeatmap
	}

	return
}

func (r *EventTicketCategoryRepositoryImpl) FindNAvailableSeatAfterSectorRowColumn(ctx context.Context, tx pgx.Tx, eventId, sectorId string, seatCount, seatRow, seatColumn int) (seats []entity.EventVenueSector, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT *
	FROM (
		SELECT 
			vssm.id, 
			vssm.seat_row, 
			vssm.seat_column, 
			vssm.seat_row_label, 
			CASE 
				WHEN vssm.label != evssm.label THEN evssm.label
				ELSE vssm.label
			END AS seat_final_label,
			CASE 
				WHEN vssm.status != evssm.status THEN 
					CASE 
						WHEN evssm.status IN ('PREBOOK', 'COMPLIMENT') THEN 'UNAVAILABLE'
						ELSE evssm.status
					END 
				ELSE vssm.status
			END AS seat_final_status
		FROM venue_sector_seatmap_matrix vssm 
		LEFT JOIN event_venue_sector_seatmap_matrix evssm 
			ON vssm.sector_id = evssm.sector_id 
			AND vssm.seat_row = evssm.seat_row 
			AND vssm.seat_column = evssm.seat_column
			AND evssm.event_id = $1
		WHERE vssm.sector_id = $2
		AND (vssm.seat_row, vssm.seat_column) > ($3, $4)
	) sub
	WHERE seat_final_status = $5
	ORDER BY seat_row ASC, seat_column ASC 
	LIMIT $6
`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, eventId, sectorId, seatRow, seatColumn, lib.SeatmapStatusAvailable, seatCount)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, eventId, sectorId, eventId, sectorId, seatRow, seatColumn, lib.SeatmapStatusAvailable, seatCount)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var sectorSeatmap entity.EventVenueSector
		rows.Scan(
			&sectorSeatmap.ID,
			&sectorSeatmap.SeatRow,
			&sectorSeatmap.SeatColumn,
			&sectorSeatmap.SeatRowLabel,
			&sectorSeatmap.Label,
			&sectorSeatmap.Status,
		)

		seats = append(seats, sectorSeatmap)
	}

	log.Info().Int("RequestSeat", seatCount).Int("Result", len(seats)).Msg("Checking seats")

	if len(seats) != seatCount {
		return seats, &lib.ErrorSeatAvailableSeatNotMatcheWithRequestSeats
	}

	return
}
