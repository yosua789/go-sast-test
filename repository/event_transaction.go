package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/dto"
	"assist-tix/entity"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"errors"

	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

type EventTransactionRepository interface {
	CreateTransaction(ctx context.Context, tx pgx.Tx, eventId, eventTicketCategoryId string, req model.EventTransaction) (res model.EventTransaction, err error)
	IsEmailAlreadyBookEvent(ctx context.Context, tx pgx.Tx, eventId, email string) (id string, err error)
	FindByOrderNumber(ctx context.Context, tx pgx.Tx, orderNumber string) (res model.EventTransaction, err error)
	MarkTransactionAsSuccess(ctx context.Context, tx pgx.Tx, transactionID string, successTime time.Time, pgOrderID string) (res model.EventTransaction, err error)
	UpdatePaymentAdditionalInformation(ctx context.Context, tx pgx.Tx, transactionID, vaNo string) (err error)
	FindById(ctx context.Context, tx pgx.Tx, transactionID string) (resData dto.OrderDetails, err error)
	FindTransactionDetailByTransactionId(ctx context.Context, tx pgx.Tx, transactionID string) (res entity.EventTransaction, err error)
	MarkTransactionAsFailed(ctx context.Context, tx pgx.Tx, transactionID string, pgOrderID string) (res model.EventTransaction, err error)
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

	query := `
		INSERT INTO event_transactions (
		order_number,
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

		created_at,
		pg_additional_fee
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), $16) RETURNING id, created_at`

	if tx != nil {
		err = tx.QueryRow(ctx, query,
			req.OrderNumber,
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
			req.PGAdditionalFee, // Additional fee for payment gateway
		).Scan(&req.ID, &req.CreatedAt)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query,
			req.OrderNumber,
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
			req.PGAdditionalFee, // Additional fee for payment gateway
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

func (r *EventTransactionRepositoryImpl) FindByOrderNumber(ctx context.Context, tx pgx.Tx, orderNumber string) (res model.EventTransaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `
	SELECT id 
	FROM event_transactions 
	WHERE order_number = $1 LIMIT 1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, orderNumber).Scan(&res.ID)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, orderNumber).Scan(&res.ID)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EventTransaction{}, nil
		}
		return
	}

	return
}

func (r *EventTransactionRepositoryImpl) MarkTransactionAsSuccess(ctx context.Context, tx pgx.Tx, transactionID string, successTime time.Time, pgOrderID string) (res model.EventTransaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE event_transactions SET transaction_status = $1, paid_at = $2, pg_order_id = $3 WHERE id = $4 RETURNING id, created_at`
	if tx != nil {
		err = tx.QueryRow(ctx, query, lib.EventTransactionStatusSuccess, successTime, pgOrderID, transactionID).Scan(&res.ID, &res.CreatedAt)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, lib.EventTransactionStatusSuccess, successTime, pgOrderID, transactionID).Scan(&res.ID, &res.CreatedAt)
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

func (r *EventTransactionRepositoryImpl) MarkTransactionAsFailed(ctx context.Context, tx pgx.Tx, transactionID string, pgOrderID string) (res model.EventTransaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	currentTime := time.Now()
	defer cancel()

	query := `UPDATE event_transactions SET transaction_status = $1,  pg_order_id = $2,updated_at = $3 WHERE id = $4 RETURNING id, created_at`
	if tx != nil {
		err = tx.QueryRow(ctx, query, lib.EventTransactionStatusFailed, pgOrderID, currentTime, transactionID).Scan(&res.ID, &res.CreatedAt)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, lib.EventTransactionStatusFailed, pgOrderID, currentTime, transactionID).Scan(&res.ID, &res.CreatedAt)
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

func (r *EventTransactionRepositoryImpl) UpdatePaymentAdditionalInformation(ctx context.Context, tx pgx.Tx, transactionID, vaNo string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE event_transactions SET payment_additional_information = $1 WHERE id = $2`
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

func (r *EventTransactionRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, transactionID string) (resData dto.OrderDetails, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	var res entity.OrderDetails
	defer cancel()
	// return grand total, event time , venue place, ticket type , transaction quantity, transaction deadline, transaction status, virtual account number
	query := `
	SELECT 
	e.name,
	e.event_time,
	v.name as venue_name,
	et.payment_expired_at,
	et.transaction_status,
	et.payment_additional_information, 
	et.payment_method,
	COUNT(eti.id) AS item_count,
	et.grand_total,
	et.total_admin_fee,
	et.total_tax,
	et.total_price,
	v.country,
	v.city,
	et.pg_additional_fee
	FROM event_transactions et
	JOIN events e ON et.event_id = e.id
	JOIN venues v ON e.venue_id = v.id
	LEFT JOIN event_transaction_items eti ON et.id = eti.transaction_id
	WHERE et.id = $1
	GROUP BY 
	e.name, e.event_time, v.name,
	et.payment_expired_at, et.transaction_status,
	et.payment_additional_information, et.payment_method,
	et.grand_total, et.total_admin_fee, et.total_tax, et.total_price, 
	v.country, v.city, et.pg_additional_fee
	LIMIT 1;
	`
	if tx != nil {
		err = tx.QueryRow(ctx, query, transactionID).Scan(
			&res.EventName,
			&res.EventTime,
			&res.VenueName,
			&res.TransactionDeadline,
			&res.TransactionStatus,
			&res.PaymentAdditionalInfo, // e.g. VA Number, QR Code
			&res.PaymentMethod,
			&res.TransactionQuantity,
			&res.GrandTotal,
			&res.TotalAdminFee,
			&res.TotalTax,
			&res.TotalPrice,
			&res.Country,
			&res.City,
			&res.PGAdditionalFee, // Additional fee for payment gateway
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, transactionID).Scan(
			&res.EventName,
			&res.EventTime,
			&res.VenueName,
			&res.TransactionDeadline,
			&res.TransactionStatus,
			&res.PaymentAdditionalInfo, // e.g. VA Number, QR Code
			&res.PaymentMethod,
			&res.TransactionQuantity,
			&res.GrandTotal,
			&res.TotalAdminFee,
			&res.TotalTax,
			&res.TotalPrice,
			&res.Country,
			&res.City,
			&res.PGAdditionalFee, // Additional fee for payment gateway
		)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.OrderDetails{}, &lib.ErrorTransactionDetailsNotFound
		}
		return
	}

	var resAdditionalPayment []entity.AdditionalPaymentInfo
	queryAdditionalPayment := `
	SELECT eaf.name, eaf.is_tax, eaf.is_percentage, eaf.value
	FROM event_additional_fees eaf
	JOIN event_transactions et ON eaf.event_id = et.event_id
	WHERE et.id = $1
	`
	if tx != nil {
		rows, err := tx.Query(ctx, queryAdditionalPayment, transactionID)
		if err != nil {
			return dto.OrderDetails{}, err
		}
		defer rows.Close()
		for rows.Next() {
			var additionalPayment entity.AdditionalPaymentInfo
			if err := rows.Scan(&additionalPayment.Name, &additionalPayment.IsTax, &additionalPayment.IsPercentage, &additionalPayment.Value); err != nil {
				return dto.OrderDetails{}, err
			}
			resAdditionalPayment = append(resAdditionalPayment, additionalPayment)
		}
	} else {
		rows, err := r.WrapDB.Postgres.Query(ctx, queryAdditionalPayment, transactionID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to query additional payment information")
			return dto.OrderDetails{}, err
		}
		defer rows.Close()
		for rows.Next() {
			var additionalPayment entity.AdditionalPaymentInfo
			if err := rows.Scan(&additionalPayment.Name, &additionalPayment.IsTax, &additionalPayment.IsPercentage, &additionalPayment.Value); err != nil {
				return dto.OrderDetails{}, err
			}
			resAdditionalPayment = append(resAdditionalPayment, additionalPayment)
		}
	}
	for i, obj := range resAdditionalPayment {
		if obj.IsPercentage {
			resAdditionalPayment[i].Value = (float64(res.TotalPrice) * obj.Value) / 100
			resAdditionalPayment[i].IsPercentage = false
		} else {
			resAdditionalPayment[i].Value = obj.Value
		}
	}
	resData = dto.OrderDetails{
		EventName:             res.EventName,
		VenueName:             res.VenueName,
		EventTime:             res.EventTime,
		TransactionDeadline:   res.TransactionDeadline,
		TransactionStatus:     res.TransactionStatus,
		PaymentMethod:         res.PaymentMethod,
		PaymentAdditionalInfo: res.PaymentAdditionalInfo, // e.g. VA Number, QR Code
		GrandTotal:            res.GrandTotal,
		TotalAdminFee:         res.TotalAdminFee,
		TotalTax:              res.TotalTax,
		TotalPrice:            res.TotalPrice,
		TransactionQuantity:   res.TransactionQuantity,
		Country:               res.Country, // Assuming country and city are not available in this query
		City:                  res.City,
		AdditionalPayment:     resAdditionalPayment,
	}

	return
}

func (r *EventTransactionRepositoryImpl) FindTransactionDetailByTransactionId(ctx context.Context, tx pgx.Tx, transactionID string) (res entity.EventTransaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT
		et.id,
		et.order_number,
		et.transaction_status,
		et.transaction_status_information,
		et.payment_expired_at,
		et.paid_at,
		et.total_price,
		et.tax_percentage,
		et.total_tax,
		et.admin_fee_percentage,
		et.total_admin_fee,
		et.grand_total,
		et.full_name,
		et.email,
		et.is_compliment,

		pm.id as payment_method_id,
		pm.name as payment_method_name,
		pm.is_paused as payment_method_is_paused,
		pm.pause_message as payment_method_pause_message,
		pm.paused_at as payment_method_paused_at,
		pm.payment_type as payment_method_type,
		pm.payment_group as payment_method_group,
		pm.payment_code as payment_method_code,
		pm.payment_channel as payment_method_channel,

		et.event_id,
		e.name as event_name,
		e.description as event_description,
		e.banner_filename as event_banner_filename,
		e.event_time as event_time,
		e.publish_status as event_publish_status,
		e.additional_information as event_additional_information,
		e.start_sale_at as event_start_sale_at,
		e.end_sale_at as event_end_sale_at,
		e.created_at as event_created_at,
		e.updated_at as event_updated_at,

		o.id as organizer_id,
		o.name as organizer_name,
		o.slug as organizer_slug,
		o.logo as organizer_logo,

		v.id as venue_id,
		v.venue_type as venue_type,
		v.name as venue_name,
		v.country as venue_country,
		v.city as venue_city,
		v.capacity as venue_capacity,

		etc.id as event_ticket_category_id,
		etc.name as event_ticket_category_name,
		etc.description as event_ticket_category_description,
		etc.price as event_ticket_category_price,
		etc.total_stock as event_ticket_category_total_stock,
		etc.total_public_stock as event_ticket_category_total_public_stock,
		etc.public_stock as event_ticket_category_public_stock,
		etc.total_compliment_stock as event_ticket_category_total_compliment_stock,
		etc.compliment_stock as event_ticket_category_compliment_stock,
		etc.code as event_ticket_category_code,
		etc.entrance as event_ticket_category_entrance,

		vs.id as venue_sector_id,
		vs.name as venue_sector_name,
		vs.has_seatmap as venue_sector_has_seatmap,
		vs.sector_color as venue_sector_color,
		vs.area_code as venue_sector_area_code

	FROM event_transactions et
	LEFT JOIN payment_methods pm
		ON et.payment_method = pm.payment_code
	LEFT JOIN events e
		ON et.event_id = e.id
	LEFT JOIN organizers o
		ON e.organizer_id = o.id
	LEFT JOIN venues v
		ON e.venue_id = v.id
	LEFT JOIN event_ticket_categories etc
		ON et.event_ticket_category_id = etc.id
	LEFT JOIN venue_sectors vs
		ON etc.venue_sector_id = vs.id
	WHERE et.id = $1
	LIMIT 1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, transactionID).Scan(
			&res.ID,
			&res.OrderNumber,
			&res.Status,
			&res.StatusInformation,
			&res.PaymentExpiredAt,
			&res.PaidAt,
			&res.TotalPrice,
			&res.TaxPercentage,
			&res.TotalTax,
			&res.AdminFeePercentage,
			&res.TotalAdminFee,
			&res.GrandTotal,
			&res.Fullname,
			&res.Email,
			&res.IsCompliment,

			&res.PaymentMethod.ID,
			&res.PaymentMethod.Name,
			&res.PaymentMethod.IsPaused,
			&res.PaymentMethod.PauseMessage,
			&res.PaymentMethod.PausedAt,
			&res.PaymentMethod.PaymentType,
			&res.PaymentMethod.PaymentGroup,
			&res.PaymentMethod.PaymentCode,
			&res.PaymentMethod.PaymentChannel,

			&res.Event.ID,
			&res.Event.Name,
			&res.Event.Description,
			&res.Event.Banner,
			&res.Event.EventTime,
			&res.Event.PublishStatus,
			&res.Event.AdditionalInformation,
			&res.Event.StartSaleAt,
			&res.Event.EndSaleAt,
			&res.Event.CreatedAt,
			&res.Event.UpdatedAt,

			&res.Event.Organizer.ID,
			&res.Event.Organizer.Name,
			&res.Event.Organizer.Slug,
			&res.Event.Organizer.Logo,

			&res.Event.Venue.ID,
			&res.Event.Venue.VenueType,
			&res.Event.Venue.Name,
			&res.Event.Venue.Country,
			&res.Event.Venue.City,
			&res.Event.Venue.Capacity,

			&res.TicketCategory.ID,
			&res.TicketCategory.Name,
			&res.TicketCategory.Description,
			&res.TicketCategory.Price,
			&res.TicketCategory.TotalStock,
			&res.TicketCategory.TotalPublicStock,
			&res.TicketCategory.PublicStock,
			&res.TicketCategory.TotalComplimentStock,
			&res.TicketCategory.ComplimentStock,
			&res.TicketCategory.Code,
			&res.TicketCategory.Entrance,

			&res.VenueSector.ID,
			&res.VenueSector.Name,
			&res.VenueSector.HasSeatmap,
			&res.VenueSector.SectorColor,
			&res.VenueSector.AreaCode,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, transactionID).Scan(
			&res.ID,
			&res.OrderNumber,
			&res.Status,
			&res.StatusInformation,
			&res.PaymentExpiredAt,
			&res.PaidAt,
			&res.TotalPrice,
			&res.TaxPercentage,
			&res.TotalTax,
			&res.AdminFeePercentage,
			&res.TotalAdminFee,
			&res.GrandTotal,
			&res.Fullname,
			&res.Email,
			&res.IsCompliment,

			&res.PaymentMethod.ID,
			&res.PaymentMethod.Name,
			&res.PaymentMethod.IsPaused,
			&res.PaymentMethod.PauseMessage,
			&res.PaymentMethod.PausedAt,
			&res.PaymentMethod.PaymentType,
			&res.PaymentMethod.PaymentGroup,
			&res.PaymentMethod.PaymentCode,
			&res.PaymentMethod.PaymentChannel,

			&res.Event.ID,
			&res.Event.Name,
			&res.Event.Description,
			&res.Event.Banner,
			&res.Event.EventTime,
			&res.Event.PublishStatus,
			&res.Event.AdditionalInformation,
			&res.Event.StartSaleAt,
			&res.Event.EndSaleAt,
			&res.Event.CreatedAt,
			&res.Event.UpdatedAt,

			&res.Event.Organizer.ID,
			&res.Event.Organizer.Name,
			&res.Event.Organizer.Slug,
			&res.Event.Organizer.Logo,

			&res.Event.Venue.ID,
			&res.Event.Venue.VenueType,
			&res.Event.Venue.Name,
			&res.Event.Venue.Country,
			&res.Event.Venue.City,
			&res.Event.Venue.Capacity,

			&res.TicketCategory.ID,
			&res.TicketCategory.Name,
			&res.TicketCategory.Description,
			&res.TicketCategory.Price,
			&res.TicketCategory.TotalStock,
			&res.TicketCategory.TotalPublicStock,
			&res.TicketCategory.PublicStock,
			&res.TicketCategory.TotalComplimentStock,
			&res.TicketCategory.ComplimentStock,
			&res.TicketCategory.Code,
			&res.TicketCategory.Entrance,

			&res.VenueSector.ID,
			&res.VenueSector.Name,
			&res.VenueSector.HasSeatmap,
			&res.VenueSector.SectorColor,
			&res.VenueSector.AreaCode,
		)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, &lib.ErrorTransactionDetailsNotFound
		}

		return res, err
	}

	return
}
