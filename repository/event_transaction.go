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

type EventTransactionRepository interface {
	CreateTransaction(ctx context.Context, tx pgx.Tx, req model.EventTransaction) (res model.EventTransaction, err error)
}

type EventTransactionRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewEventTransactionRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) EventTransactionRepository {
	return &EventTransactionRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *EventTransactionRepositoryImpl) CreateTransaction(ctx context.Context, tx pgx.Tx, req model.EventTransaction) (res model.EventTransaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO event_transactions (
		invoice_number,
		transaction_status,
		transaction_status_information, 

		payment_method,
		payment_channel,
		payment_expired_at,

		total_price, 
		tax_percentage,
		total_tax,
		admin_fee_percentage,
		total_admin_fee,
		grand_total,

		full_name,
		email,
		phone_number,

		is_compliment,

		created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW()) RETURNING id`

	if tx != nil {
		err = tx.QueryRow(ctx, query,
			req.InvoiceNumber,
			req.Status,
			req.StatusInformation,
			req.PaymentMethod,
			req.PaymentChannel,
			req.PaymentExpiredAt,
			req.TotalPrice,
			req.TaxPercentage,
			req.TotalTax,
			req.AdminFeePercentage,
			req.TotalAdminFee,
			req.GrandTotal,
			req.FullName,
			req.Email,
			req.PhoneNumber,
			req.IsCompliment,
		).Scan(&req.ID)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query,
			req.InvoiceNumber,
			req.Status,
			req.StatusInformation,
			req.PaymentMethod,
			req.PaymentChannel,
			req.PaymentExpiredAt,
			req.TotalPrice,
			req.TaxPercentage,
			req.TotalTax,
			req.AdminFeePercentage,
			req.TotalAdminFee,
			req.GrandTotal,
			req.FullName,
			req.Email,
			req.PhoneNumber,
			req.IsCompliment,
		).Scan(&req.ID)
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				err = &lib.ErrorFailedToCreateTransaction
			}
		}

		return
	}

	res = req

	return
}
