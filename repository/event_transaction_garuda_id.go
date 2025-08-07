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
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type EventTransactionGarudaIDRepository interface {
	Create(ctx context.Context, tx pgx.Tx, eventID string, garudaID string) (err error)
	GetEventGarudaID(ctx context.Context, tx pgx.Tx, eventID string, garudaID string) (res model.EventTransactionGarudaID, err error)
	CreateBatch(ctx context.Context, tx pgx.Tx, payloads dto.BulkGarudaIDRequest) (err error)
	CreateGarudaIdBooks(ctx context.Context, tx pgx.Tx, eventId string, garudaIds ...string) (err error)
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
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()
	query := `SELECT id, event_id, garuda_id, created_at FROM event_transaction_garuda_id_books WHERE event_id = $1 AND garuda_id = $2 LIMIT 1`

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

	query := `INSERT INTO event_transaction_garuda_id_books (event_id, garuda_id) VALUES ($1, $2)`

	if tx != nil {
		_, err = tx.Exec(ctx, query, eventID, garudaID)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, eventID, garudaID)
	}

	return err
}

func (r *EventTransactionGarudaIDRepositoryImpl) CreateGarudaIdBooks(ctx context.Context, tx pgx.Tx, eventId string, garudaIds ...string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	if len(garudaIds) == 0 {
		return nil
	}

	query := `INSERT INTO event_transaction_garuda_id_books (
		event_id,
		garuda_id,
		created_at
	) VALUES `

	var args []interface{}
	var placeholders []string

	for i, garudaId := range garudaIds {
		base := i * 2
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, NOW())",
			base+1, base+2))

		args = append(args,
			eventId,
			garudaId,
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
				err = &lib.ErrorGarudaIDAlreadyUsed
			}
		}

		return err
	}

	return nil
}

func (r *EventTransactionGarudaIDRepositoryImpl) CreateBatch(ctx context.Context, tx pgx.Tx, payload dto.BulkGarudaIDRequest) error {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	if len(payload.GarudaIDs) == 0 {
		return nil // No data to insert
	}

	rows := make([][]interface{}, 0, len(payload.GarudaIDs))
	for _, garudaID := range payload.GarudaIDs {
		// Explicit cast to string to avoid COPY encoding error
		rows = append(rows, []interface{}{string(payload.EventID), string(garudaID)})
	}

	columns := []string{"event_id", "garuda_id"}

	var (
		copyCount int64
		err       error
	)

	if tx != nil {
		copyCount, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"event_transaction_garuda_id_books"},
			columns,
			pgx.CopyFromRows(rows),
		)
	} else {
		copyCount, err = r.WrapDB.Postgres.CopyFrom(
			ctx,
			pgx.Identifier{"event_transaction_garuda_id_books"},
			columns,
			pgx.CopyFromRows(rows),
		)
	}

	if err != nil {
		return &lib.ErrorInternalServer
	}

	if int(copyCount) != len(rows) {
		return &lib.ErrorInternalServer
	}

	return nil
}
