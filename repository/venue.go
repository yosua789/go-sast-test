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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

type VenueRepository interface {
	Create(ctx context.Context, tx pgx.Tx, venue model.Venue) (id string, err error)
	FindAll(ctx context.Context, tx pgx.Tx) (res []model.Venue, err error)
	FindById(ctx context.Context, tx pgx.Tx, venueId string) (venue model.Venue, err error)
	Update(ctx context.Context, tx pgx.Tx, venue model.Venue) (err error)
	SoftDelete(ctx context.Context, tx pgx.Tx, venueId string) (err error)
}

type VenueRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewVenueRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) VenueRepository {
	return &VenueRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *VenueRepositoryImpl) Create(ctx context.Context, tx pgx.Tx, venue model.Venue) (id string, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO venues (venue_type, name, country, city, status, capacity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`

	if tx != nil {
		_, err = tx.Exec(ctx, query, venue.VenueType, venue.Name, venue.Country, venue.City, venue.Status, venue.Capacity)
	} else {
		_, err = r.WrapDB.Postgres.Conn.Exec(ctx, query, venue.VenueType, venue.Name, venue.Country, venue.City, venue.Status, venue.Capacity)
	}

	return
}

func (r *VenueRepositoryImpl) FindAll(ctx context.Context, tx pgx.Tx) (res []model.Venue, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT id, venue_type, name, country, city, status, capacity, created_at, updated_at FROM venues WHERE deleted_at IS NULL`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query)
	} else {
		rows, err = r.WrapDB.Postgres.Conn.Query(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var venue model.Venue
		rows.Scan(
			&venue.ID,
			&venue.VenueType,
			&venue.Name,
			&venue.Country,
			&venue.City,
			&venue.Status,
			&venue.Capacity,
			&venue.CreatedAt,
			&venue.UpdatedAt,
		)

		res = append(res, venue)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("FindAll venue error")
		return res, rows.Err()
	}

	return
}

func (r *VenueRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, venueId string) (venue model.Venue, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT id, venue_type, name, country, city, capacity, status, created_at, updated_at FROM venues WHERE id = $1 AND deleted_at IS NULL`

	if tx != nil {
		err = tx.QueryRow(ctx, query, venueId).Scan(
			&venue.ID,
			&venue.VenueType,
			&venue.Name,
			&venue.Country,
			&venue.City,
			&venue.Capacity,
			&venue.Status,
			&venue.CreatedAt,
			&venue.UpdatedAt,
		)
	} else {
		err = r.WrapDB.Postgres.Conn.QueryRow(ctx, query, venueId).Scan(
			&venue.ID,
			&venue.VenueType,
			&venue.Name,
			&venue.Country,
			&venue.City,
			&venue.Capacity,
			&venue.Status,
			&venue.CreatedAt,
			&venue.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return venue, &lib.ErrorVenueNotFound
		}
		return venue, err
	}

	return
}

func (r *VenueRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, venue model.Venue) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE venues SET
		venue_type = COALESCE($1, venue_type), 
		name = COALESCE($2, name), 
		country = COALESCE($3, country),
		city = COALESCE($4, city),
		status = COALESCE($5, status), 
		capacity = COALESCE($6, capacity), 
		updated_at = CURRENT_TIMESTAMP
		WHERE id = $7 AND deleted_at IS NULL`

	var cmdTag pgconn.CommandTag

	if tx != nil {
		cmdTag, err = tx.Exec(ctx, query, venue.VenueType, venue.Name, venue.Country, venue.City, venue.Status, venue.Capacity, venue.ID)
	} else {
		cmdTag, err = r.WrapDB.Postgres.Conn.Exec(ctx, query, venue.VenueType, venue.Name, venue.Country, venue.City, venue.Status, venue.Capacity, venue.ID)
	}

	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok {
			if pgErr.Code == "23505" {
				return &lib.ErrorVenueNameConflict
			}
		}
		return err
	}

	// TODO: Check if the update wasn't successful
	if cmdTag.RowsAffected() == 0 {
	}

	return
}

func (r *VenueRepositoryImpl) SoftDelete(ctx context.Context, tx pgx.Tx, venueId string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE venues SET
		deleted_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL`

	// var cmdTag pgconn.CommandTag
	if tx != nil {
		_, err = tx.Exec(ctx, query, venueId)
	} else {
		_, err = r.WrapDB.Postgres.Conn.Exec(ctx, query, venueId)
	}

	return
}
