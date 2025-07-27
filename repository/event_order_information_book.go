package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/lib"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type EventOrderInformationBookRepository interface {
	CreateOrderInformation(ctx context.Context, tx pgx.Tx, eventId, email, fullname string) (id int, err error)
	UpdateTransactionIdByID(ctx context.Context, tx pgx.Tx, id int, transactionId string) (err error)
}

type EventOrderInformationBookRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewEventOrderInformationBookRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) EventOrderInformationBookRepository {
	return &EventOrderInformationBookRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *EventOrderInformationBookRepositoryImpl) CreateOrderInformation(ctx context.Context, tx pgx.Tx, eventId, email, fullname string) (id int, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO event_order_information_books (
		event_id,
		email,
		full_name,
		created_at
	) VALUES ($1, $2, $3, NOW()) RETURNING id`

	if tx != nil {
		err = tx.QueryRow(ctx, query, eventId, email, fullname).Scan(&id)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, eventId, email, fullname).Scan(&id)
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				err = &lib.ErrorEmailIsAlreadyBooked
			}
		}
		return
	}

	return
}

func (r *EventOrderInformationBookRepositoryImpl) UpdateTransactionIdByID(ctx context.Context, tx pgx.Tx, id int, transactionId string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE event_order_information_books SET event_transaction_id = $1 WHERE id = $2`

	if tx != nil {
		_, err = tx.Exec(ctx, query, transactionId, id)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, transactionId, id)
	}

	if err != nil {
		return
	}

	return
}
