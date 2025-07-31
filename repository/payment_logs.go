package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PaymentLogRepository interface {
	Create(ctx context.Context, tx pgx.Tx, paymentLog model.PaymentLog) (res model.PaymentLog, err error)
}

type PaymentLogRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewPaymentLogRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) PaymentLogRepository {
	return &PaymentLogRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}
func (r *PaymentLogRepositoryImpl) Create(ctx context.Context, tx pgx.Tx, paymentLog model.PaymentLog) (res model.PaymentLog, err error) {
	res = paymentLog
	// Create a new payment log in the database
	query := `INSERT INTO payment_logs ( header, body, response,  error_response, endpoint_path, error_code)
			 VALUES ($1, $2, $3, $4, $5, $6) returning id`
	if tx != nil {
		err = tx.QueryRow(ctx, query,
			paymentLog.Header,
			paymentLog.Body,
			paymentLog.Response,
			paymentLog.ErrorResponse,
			paymentLog.Path,
			paymentLog.ErrorCode).Scan(&res.ID)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query,
			paymentLog.Header,
			paymentLog.Body,
			paymentLog.Response,
			paymentLog.ErrorResponse,
			paymentLog.Path,
			paymentLog.ErrorCode).Scan(&res.ID)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				err = &lib.ErrorCreatePaymentLog
			}
		}

		return
	}
	// If the insert was successful, return nil
	return
}
