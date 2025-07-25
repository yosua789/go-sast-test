package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"database/sql"
	"errors"

	"github.com/rs/zerolog/log"
)

type EventTransactionGarudaIDRepository interface {
	Create(ctx context.Context, eventID string, garudaID string) error
	GetEventGarudaID(ctx context.Context, eventID string, garudaID string) error
}

type EventTransactionGarudaIDRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewEventTransactionGarudaIDRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) EventTransactionGarudaIDRepository {
	return &EventTransactionGarudaIDRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *EventTransactionGarudaIDRepositoryImpl) GetEventGarudaID(ctx context.Context, eventID string, garudaID string) error {
	query := `SELECT * FROM event_transaction_garuda_id WHERE event_id = $1 AND garuda_id = $2`
	rows, err := r.WrapDB.Postgres.Query(ctx, query, eventID, garudaID)
	if errors.Is(err, sql.ErrNoRows) {
		return &lib.ErrorEventNotFound
	}
	defer rows.Close()

	var results []model.EventTransactionGarudaID
	for rows.Next() {
		var result model.EventTransactionGarudaID
		if err := rows.Scan(&result); err != nil {
			log.Error().Err(err).Msg("failed to scan row")
			return &lib.ErrorInternalServer
		}
		results = append(results, result)
	}
	if len(results) > 0 {
		return &lib.ErrorGarudaIDAlreadyUsed
	}
	return nil
}
func (r *EventTransactionGarudaIDRepositoryImpl) Create(ctx context.Context, eventID string, garudaID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO event_transaction_garuda_id (event_id, garuda_id) VALUES ($1, $2)`
	_, err := r.WrapDB.Postgres.Exec(ctx, query, eventID, garudaID)
	return err
}
