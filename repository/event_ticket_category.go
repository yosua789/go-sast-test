package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type EventTicketCategoryRepository interface {
	Create(ctx context.Context, tx pgx.Tx, ticketCategory model.EventTicketCategory) (err error)
	FindByEventId(ctx context.Context, tx pgx.Tx, eventId string) (ticketCategories []model.EventTicketCategory, err error)
	FindByIdAndEventId(ctx context.Context, tx pgx.Tx, eventId string, ticketCategoryId string) (res model.EventTicketCategory, err error)
	SoftDelete(ctx context.Context, tx pgx.Tx, ticketCategoryId string) (err error)
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

	query := `INSERT INTO event_ticket_categories (event_id, name, description, price, total_stock, total_public_stock, public_stock, total_compliment_stock, compliment_stock, code, entrance, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())`

	if tx != nil {
		_, err = tx.Exec(ctx, query,
			ticketCategory.EventID,
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
		_, err = r.WrapDB.Postgres.Conn.Exec(ctx, query,
			ticketCategory.EventID,
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
		rows, err = r.WrapDB.Postgres.Conn.Query(ctx, query, eventId)
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
		err = r.WrapDB.Postgres.Conn.QueryRow(ctx, query, eventId, ticketCategoryId).Scan(
			&res.ID,
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

	// var cmdTag pgconn.CommandTag
	if tx != nil {
		_, err = tx.Exec(ctx, query, ticketCategoryId)
	} else {
		_, err = r.WrapDB.Postgres.Conn.Exec(ctx, query, ticketCategoryId)
	}

	return
}
