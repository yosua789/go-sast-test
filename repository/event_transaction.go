package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type EventTransactionRepository interface {
	CreateTransaction(ctx context.Context, tx pgx.Tx, eventId, eventTicketCategoryId string, req model.EventTransaction) (res model.EventTransaction, err error)
	IsEmailAlreadyBookEvent(ctx context.Context, tx pgx.Tx, eventId, email string) (id string, err error)
	FindByInvoiceNumber(ctx context.Context, tx pgx.Tx, invoiceNumber string) (res model.EventTransaction, err error)
	MarkTransactionAsSuccess(ctx context.Context, tx pgx.Tx, transactionID string) (res model.EventTransaction, err error)
	UpdateVANo(ctx context.Context, tx pgx.Tx, transactionID, vaNo string) (err error)
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

	err = &lib.ErrorOrderInformationIsAlreadyBook

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
		total_tax,
		total_admin_fee,
		grand_total,

		email,		
		full_name,

		is_compliment,

		created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW()) RETURNING id, created_at`

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
			// req.TaxPercentage,
			req.TotalTax,
			// req.AdminFeePercentage,
			req.TotalAdminFee,
			req.GrandTotal,
			req.Email,
			req.Fullname,
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
			// req.TaxPercentage,
			req.TotalTax,
			// req.AdminFeePercentage,
			req.TotalAdminFee,
			req.GrandTotal,
			req.Email,
			req.Fullname,
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

func (r *EventTransactionRepositoryImpl) FindByInvoiceNumber(ctx context.Context, tx pgx.Tx, invoiceNumber string) (res model.EventTransaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `
	SELECT id 
	FROM event_transactions 
	WHERE invoice_number = $1 LIMIT 1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, invoiceNumber).Scan(&res.ID)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, invoiceNumber).Scan(&res.ID)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EventTransaction{}, nil
		}
		return
	}

	return
}

func (r *EventTransactionRepositoryImpl) MarkTransactionAsSuccess(ctx context.Context, tx pgx.Tx, transactionID string) (res model.EventTransaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	currentTime := time.Now()
	defer cancel()

	query := `UPDATE event_transactions SET transaction_status = $1, updated_at = $2 WHERE id = $3 RETURNING id, created_at`
	if tx != nil {
		err = tx.QueryRow(ctx, query, lib.EventTransactionStatusSuccess, currentTime, transactionID).Scan(&res.ID, &res.CreatedAt)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, lib.EventTransactionStatusSuccess, currentTime, transactionID).Scan(&res.ID, &res.CreatedAt)
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				err = &lib.ErrorFailedToMarkTransactionAsSuccess
			}
		}

		return
	}

	return
}

func (r *EventTransactionRepositoryImpl) UpdateVANo(ctx context.Context, tx pgx.Tx, transactionID, vaNo string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE event_transactions SET virtual_account_number = $1 WHERE id = $2`
	if tx != nil {
		_, err = tx.Exec(ctx, query, vaNo, transactionID)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, vaNo, transactionID)
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				err = &lib.ErrorFailedToUpdateVANo
			}
		}

		return
	}

	return
}
