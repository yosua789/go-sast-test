package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/entity"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

type VenueSectorRepository interface {
	FindByVenueId(ctx context.Context, tx pgx.Tx, venueId string) (sectors []model.VenueSector, err error)
	FindById(ctx context.Context, tx pgx.Tx, sectorId string) (venue model.VenueSector, err error)
	FindVenueSectorById(ctx context.Context, tx pgx.Tx, sectorId string) (venueSector entity.VenueSector, err error)
}

type VenueSectorRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewVenueSectorRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) VenueSectorRepository {
	return &VenueSectorRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *VenueSectorRepositoryImpl) FindVenueSectorById(ctx context.Context, tx pgx.Tx, sectorId string) (venueSector entity.VenueSector, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT
		vs.id, 
		vs.name,
		vs.sector_row,
		vs.sector_column,
		vs.capacity,
		vs.has_seatmap,
		vs.sector_color,
		vs.area_code,

		v.id as venue_id,
		v.venue_type as venue_type,
		v.name as venue_name,
		v.country as venue_country,
		v.city as venue_city,
		v.capacity as venue_capacity		

	FROM venue_sectors vs
	INNER JOIN venues v
		ON vs.venue_id = v.id
		AND vs.deleted_at is null
	WHERE vs.id = $1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, sectorId).Scan(
			&venueSector.ID,
			&venueSector.Name,
			&venueSector.SectorRow,
			&venueSector.SectorColumn,
			&venueSector.Capacity,
			&venueSector.HasSeatmap,
			&venueSector.SectorColor,
			&venueSector.AreaCode,
			&venueSector.Venue.ID,
			&venueSector.Venue.VenueType,
			&venueSector.Venue.Name,
			&venueSector.Venue.Country,
			&venueSector.Venue.City,
			&venueSector.Venue.Capacity,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, sectorId).Scan(
			&venueSector.ID,
			&venueSector.Name,
			&venueSector.SectorRow,
			&venueSector.SectorColumn,
			&venueSector.Capacity,
			&venueSector.HasSeatmap,
			&venueSector.SectorColor,
			&venueSector.AreaCode,
			&venueSector.Venue.ID,
			&venueSector.Venue.VenueType,
			&venueSector.Venue.Name,
			&venueSector.Venue.Country,
			&venueSector.Venue.City,
			&venueSector.Venue.Capacity,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return venueSector, &lib.ErrorVenueSectorNotFound
		}
	}

	return
}

func (r *VenueSectorRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, sectorId string) (venueSector model.VenueSector, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT id, venue_id, name, sector_row, sector_column, capacity, is_active, has_seatmap, sector_color, area_code, created_at, updated_at FROM venue_sectors WHERE id = $1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, sectorId).Scan(
			&venueSector.ID,
			&venueSector.VenueID,
			&venueSector.Name,
			&venueSector.SectorRow,
			&venueSector.SectorColumn,
			&venueSector.Capacity,
			&venueSector.IsActive,
			&venueSector.HasSeatmap,
			&venueSector.SectorColor,
			&venueSector.AreaCode,
			&venueSector.CreatedAt,
			&venueSector.UpdatedAt,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, sectorId).Scan(
			&venueSector.ID,
			&venueSector.VenueID,
			&venueSector.Name,
			&venueSector.SectorRow,
			&venueSector.SectorColumn,
			&venueSector.Capacity,
			&venueSector.IsActive,
			&venueSector.HasSeatmap,
			&venueSector.SectorColor,
			&venueSector.AreaCode,
			&venueSector.CreatedAt,
			&venueSector.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return venueSector, &lib.ErrorVenueNotFound
		}
		return venueSector, err
	}

	return
}

func (r *VenueSectorRepositoryImpl) FindByVenueId(ctx context.Context, tx pgx.Tx, venueId string) (sectors []model.VenueSector, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT 
		id, 
		venue_id, 
		name, 
		sector_row, 
		sector_column, 
		capacity, 
		is_active, 
		has_seatmap, 
		sector_color, 
		area_code, 
		created_at, 
		updated_at
	FROM venue_sectors 
	WHERE venue_id = $1`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query, venueId)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, venueId)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var sector model.VenueSector
		rows.Scan(
			&sector.ID,
			&sector.VenueID,
			&sector.Name,
			&sector.SectorRow,
			&sector.SectorColumn,
			&sector.Capacity,
			&sector.IsActive,
			&sector.HasSeatmap,
			&sector.SectorColor,
			&sector.AreaCode,
			&sector.CreatedAt,
			&sector.UpdatedAt,
		)

		sectors = append(sectors, sector)
	}

	return
}
