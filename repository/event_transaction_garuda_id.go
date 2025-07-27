package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/dto"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

type EventTransactionGarudaIDRepository interface {
	Create(ctx context.Context, tx pgx.Tx, eventID string, garudaID string) (err error)
	GetEventGarudaID(ctx context.Context, tx pgx.Tx, eventID string, garudaID string) (res model.EventTransactionGarudaID, err error)
	CreateBatch(ctx context.Context, tx pgx.Tx, payloads dto.BulkGarudaIDRequest) (err error)
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

func (r *EventTransactionGarudaIDRepositoryImpl) GetEventGarudaID(ctx context.Context, tx pgx.Tx, eventID string, garudaID string) (res model.EventTransactionGarudaID, err error) {
	query := `SELECT id, event_id, garuda_id, created_at FROM transaction_garuda_id_books WHERE event_id = $1 AND garuda_id = $2 LIMIT 1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, eventID, garudaID).Scan(
			&res.ID,
			&res.EventID,
			&res.GarudaID,
			&res.CreatedAt,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, eventID, garudaID).Scan(
			&res.ID,
			&res.EventID,
			&res.GarudaID,
			&res.CreatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = &lib.ErrorGarudaIDNotFound
			return
		}

		return res, err
	}

	return res, nil
}

func (r *EventTransactionGarudaIDRepositoryImpl) Create(ctx context.Context, tx pgx.Tx, eventID string, garudaID string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO transaction_garuda_id_books (event_id, garuda_id) VALUES ($1, $2)`

	if tx != nil {
		_, err = tx.Exec(ctx, query, eventID, garudaID)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, eventID, garudaID)
	}

	return err
}

func (r *EventTransactionGarudaIDRepositoryImpl) CreateBatch(ctx context.Context, tx pgx.Tx, payloads dto.BulkGarudaIDRequest) error {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	var rows [][]interface{}

	for _, payload := range payloads.GarudaIDs {
		for _, garudaID := range payload {
			rows = append(rows, []interface{}{payloads.EventID, garudaID})
		}
	}

	if len(rows) == 0 {
		return nil // Nothing to insert
	}

	columns := []string{"event_id", "garuda_id"}

	var copyCount int64
	var err error

	if tx != nil {
		copyCount, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"transaction_garuda_id_books"},
			columns,
			pgx.CopyFromRows(rows),
		)
	} else {
		copyCount, err = r.WrapDB.Postgres.CopyFrom(
			ctx,
			pgx.Identifier{"transaction_garuda_id_books"},
			columns,
			pgx.CopyFromRows(rows),
		)
	}

	if err != nil {
		return err
	}

	expected := len(rows)
	if int(copyCount) != expected {
		return &lib.ErrorInternalServer
	}

	return nil
}
