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
)

type VenueSectorRepository interface {
	FindById(ctx context.Context, tx pgx.Tx, sectorId string) (venue model.VenueSector, err error)
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
