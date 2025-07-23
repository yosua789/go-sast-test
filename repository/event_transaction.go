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
	CreateTransaction(ctx context.Context, tx pgx.Tx, eventId, eventTicketCategoryId string, req model.EventTransaction) (res model.EventTransaction, err error)
	IsEmailAlreadyBookEvent(ctx context.Context, tx pgx.Tx, eventId, email string) (id string, err error)
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

func (r *EventTransactionRepositoryImpl) IsEmailAlreadyBookEvent(ctx context.Context, tx pgx.Tx, eventId, email string) (id string, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT id FROM event_transactions WHERE email = $1 AND event_id = $2 AND (transaction_status = $3 OR transaction_status = $4) LIMIT 1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, email, eventId, lib.EventTransactionStatusPending, lib.EventTransactionStatusSuccess).Scan(&id)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, email, eventId, lib.EventTransactionStatusPending, lib.EventTransactionStatusSuccess).Scan(&id)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}

		return
	}

	err = &lib.ErrorEmailIsAlreadyBooked

	return
}

func (r *EventTransactionRepositoryImpl) CreateTransaction(ctx context.Context, tx pgx.Tx, eventId, eventTicketCategoryId string, req model.EventTransaction) (res model.EventTransaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO event_transactions (
		invoice_number,
		transaction_status,
		transaction_status_information, 

		event_id,
		event_ticket_category_id,

		payment_method,
		payment_channel,
		payment_expired_at,

		total_price, 
		tax_percentage,
		total_tax,
		admin_fee_percentage,
		total_admin_fee,
		grand_total,

		email,		

		is_compliment,

		created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW()) RETURNING id, created_at`

	if tx != nil {
		err = tx.QueryRow(ctx, query,
			req.InvoiceNumber,
			req.Status,
			req.StatusInformation,
			eventId,
			eventTicketCategoryId,
			req.PaymentMethod,
			req.PaymentChannel,
			req.PaymentExpiredAt,
			req.TotalPrice,
			req.TaxPercentage,
			req.TotalTax,
			req.AdminFeePercentage,
			req.TotalAdminFee,
			req.GrandTotal,
			req.Email,
			req.IsCompliment,
		).Scan(&req.ID, &req.CreatedAt)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query,
			req.InvoiceNumber,
			req.Status,
			req.StatusInformation,
			eventId,
			eventTicketCategoryId,
			req.PaymentMethod,
			req.PaymentChannel,
			req.PaymentExpiredAt,
			req.TotalPrice,
			req.TaxPercentage,
			req.TotalTax,
			req.AdminFeePercentage,
			req.TotalAdminFee,
			req.GrandTotal,
			req.Email,
			req.IsCompliment,
		).Scan(&req.ID, &req.CreatedAt)
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
