package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/helper"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

type OrganizerRepository interface {
	Create(ctx context.Context, tx pgx.Tx, organizer model.Organizer) (id string, err error)
	FindAll(ctx context.Context, tx pgx.Tx) (res []model.Organizer, err error)
	FindById(ctx context.Context, tx pgx.Tx, organizerId string) (organizer model.Organizer, err error)
	FindByIds(ctx context.Context, tx pgx.Tx, organizerIds ...string) (res []model.Organizer, err error)
	Update(ctx context.Context, tx pgx.Tx, organizer model.Organizer) (err error)
	SoftDelete(ctx context.Context, tx pgx.Tx, organizerId string) (err error)
}

type OrganizerRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewOrganizerRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) OrganizerRepository {
	return &OrganizerRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *OrganizerRepositoryImpl) Create(ctx context.Context, tx pgx.Tx, organizer model.Organizer) (id string, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO organizers (name, slug, logo, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())`

	if tx != nil {
		_, err = tx.Exec(ctx, query, organizer.Name, organizer.Slug, organizer.Logo)
	} else {
		_, err = r.WrapDB.Postgres.Conn.Exec(ctx, query, organizer.Name, organizer.Slug, organizer.Logo)
	}

	return
}

func (r *OrganizerRepositoryImpl) FindAll(ctx context.Context, tx pgx.Tx) (res []model.Organizer, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT id, name, slug, logo, created_at, updated_at FROM organizers WHERE deleted_at IS NULL`

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
		var organizer model.Organizer
		rows.Scan(
			&organizer.ID,
			&organizer.Name,
			&organizer.Slug,
			&organizer.Logo,
			&organizer.CreatedAt,
			&organizer.UpdatedAt,
		)

		res = append(res, organizer)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("FindAll organizer error")
		return res, rows.Err()
	}

	return
}

func (r *OrganizerRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, organizerId string) (organizer model.Organizer, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT id, name, slug, logo, created_at, updated_at FROM organizers WHERE id = $1 AND deleted_at IS NULL`

	if tx != nil {
		err = tx.QueryRow(ctx, query, organizerId).Scan(
			&organizer.ID,
			&organizer.Name,
			&organizer.Slug,
			&organizer.Logo,
			&organizer.CreatedAt,
			&organizer.UpdatedAt,
		)
	} else {
		err = r.WrapDB.Postgres.Conn.QueryRow(ctx, query, organizerId).Scan(
			&organizer.ID,
			&organizer.Name,
			&organizer.Slug,
			&organizer.Logo,
			&organizer.CreatedAt,
			&organizer.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return organizer, &lib.ErrorOrganizerNotFound
		}
		return organizer, err
	}

	return
}

func (r *OrganizerRepositoryImpl) FindByIds(ctx context.Context, tx pgx.Tx, organizerIds ...string) (res []model.Organizer, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := fmt.Sprintf(`SELECT id, name, slug, logo, created_at, updated_at FROM organizers WHERE id IN (%s) AND deleted_at IS NULL`, helper.JoinArrayToQuotedString(organizerIds, ","))

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
		var organizer model.Organizer
		rows.Scan(
			&organizer.ID,
			&organizer.Name,
			&organizer.Slug,
			&organizer.Logo,
			&organizer.CreatedAt,
			&organizer.UpdatedAt,
		)

		res = append(res, organizer)
	}

	return
}

func (r *OrganizerRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, organizer model.Organizer) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE organizers SET
		name = COALESCE($1, name), 
		slug = COALESCE($2, slug),
		logo = COALESCE($3, logo),
		updated_at = CURRENT_TIMESTAMP 
		WHERE id = $4 AND deleted_at IS NULL`

	var cmdTag pgconn.CommandTag

	if tx != nil {
		cmdTag, err = tx.Exec(ctx, query, organizer.Name, organizer.Slug, organizer.Logo, organizer.ID)
	} else {
		cmdTag, err = r.WrapDB.Postgres.Conn.Exec(ctx, query, organizer.Name, organizer.Slug, organizer.Logo, organizer.ID)
	}

	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok {
			if pgErr.Code == "23505" {
				return &lib.ErrorOrganizerNameConflict
			}
		}
		return err
	}

	// TODO: Check if the update wasn't successful
	if cmdTag.RowsAffected() == 0 {
	}

	return
}

func (r *OrganizerRepositoryImpl) SoftDelete(ctx context.Context, tx pgx.Tx, organizerId string) (err error) {

	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE organizers SET
		deleted_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL`

	// var cmdTag pgconn.CommandTag
	if tx != nil {
		_, err = tx.Exec(ctx, query, organizerId)
	} else {
		_, err = r.WrapDB.Postgres.Conn.Exec(ctx, query, organizerId)
	}

	return
}
