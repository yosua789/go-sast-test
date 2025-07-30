package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type EventTicketRepository interface {
	Create(ctx context.Context, tx pgx.Tx, eventTicket model.EventTicket) (id int, err error)
	FindById(ctx context.Context, tx pgx.Tx, id string) (res model.EventTicket, err error)
}

type EventTicketRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewEventTicketRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) EventTicketRepository {
	return &EventTicketRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *EventTicketRepositoryImpl) Create(ctx context.Context, tx pgx.Tx, eventTicket model.EventTicket) (id int, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO event_tickets (
		event_id, 
		ticket_category_id, 
		event_transaction_id, 
		ticket_owner_email,
		ticket_owner_full_name,
		ticket_owner_phone_number,
		ticket_owner_garuda_id,
		ticket_number, 
		ticket_code, 
		event_time,
		event_venue,
		event_city,
		event_country,
		sector_name,
		area_code,
		entrance,
		seat_row,
		seat_column,
		seat_label,
		is_compliment,
		additional_information,
		created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, NOW()) RETURNING id`

	if tx != nil {
		err = tx.QueryRow(ctx, query,
			eventTicket.EventID,
			eventTicket.TicketCategoryID,
			eventTicket.TransactionID,
			eventTicket.TicketOwnerEmail,
			eventTicket.TicketOwnerFullname,
			eventTicket.TicketOwnerPhoneNumber,
			eventTicket.TicketOwnerGarudaId,
			eventTicket.TicketNumber,
			eventTicket.TicketCode,
			eventTicket.EventTime,
			eventTicket.EventVenue,
			eventTicket.EventCity,
			eventTicket.EventCountry,
			eventTicket.SectorName,
			eventTicket.AreaCode,
			eventTicket.Entrance,
			eventTicket.SeatRow,
			eventTicket.SeatColumn,
			eventTicket.SeatLabel,
			eventTicket.IsCompliment,
			eventTicket.AdditionalInformation,
		).Scan(&id)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query,
			eventTicket.EventID,
			eventTicket.TicketCategoryID,
			eventTicket.TransactionID,
			eventTicket.TicketOwnerEmail,
			eventTicket.TicketOwnerFullname,
			eventTicket.TicketOwnerPhoneNumber,
			eventTicket.TicketOwnerGarudaId,
			eventTicket.TicketNumber,
			eventTicket.TicketCode,
			eventTicket.EventTime,
			eventTicket.EventVenue,
			eventTicket.EventCity,
			eventTicket.EventCountry,
			eventTicket.SectorName,
			eventTicket.AreaCode,
			eventTicket.Entrance,
			eventTicket.SeatRow,
			eventTicket.SeatColumn,
			eventTicket.SeatLabel,
			eventTicket.IsCompliment,
			eventTicket.AdditionalInformation,
		).Scan(&id)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, &lib.FailedToCreateEventTIcket
		}

		return id, err
	}

	return
}

func (r *EventTicketRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, id string) (res model.EventTicket, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `SELECT 
		id,
		event_id, 
		ticket_category_id, 
		event_transaction_id, 
		ticket_owner_email, 
		ticket_owner_full_name,
		ticket_owner_phone_number,  
		ticket_owner_garuda_id, 
		ticket_number, 
		ticket_code, 
		event_time,
		event_venue,
		event_city,
		event_country,
		sector_name,
		area_code,
		entrance,
		seat_row,
		seat_column,
		seat_label,
		is_compliment,
		additional_information,
		created_at
	FROM event_tickets
	WHERE id = $1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, id).Scan(
			&res.EventID,
			&res.TicketCategoryID,
			&res.TransactionID,
			&res.TicketOwnerEmail,
			&res.TicketOwnerFullname,
			&res.TicketOwnerPhoneNumber,
			&res.TicketOwnerGarudaId,
			&res.TicketNumber,
			&res.TicketCode,
			&res.EventTime,
			&res.EventVenue,
			&res.EventCity,
			&res.EventCountry,
			&res.SectorName,
			&res.AreaCode,
			&res.Entrance,
			&res.SeatRow,
			&res.SeatColumn,
			&res.SeatLabel,
			&res.IsCompliment,
			&res.AdditionalInformation,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, id).Scan(
			&res.EventID,
			&res.TicketCategoryID,
			&res.TransactionID,
			&res.TicketOwnerEmail,
			&res.TicketOwnerFullname,
			&res.TicketOwnerPhoneNumber,
			&res.TicketOwnerGarudaId,
			&res.TicketNumber,
			&res.TicketCode,
			&res.EventTime,
			&res.EventVenue,
			&res.EventCity,
			&res.EventCountry,
			&res.SectorName,
			&res.AreaCode,
			&res.Entrance,
			&res.SeatRow,
			&res.SeatColumn,
			&res.SeatLabel,
			&res.IsCompliment,
			&res.AdditionalInformation,
		)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, &lib.EventTicketNotFound
		}

		return
	}

	return
}
