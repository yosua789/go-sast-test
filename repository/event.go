package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/domain"
	"assist-tix/entity"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

type EventRepository interface {
	Create(ctx context.Context, tx pgx.Tx, event model.Event) (id string, err error)
	FindAll(ctx context.Context, tx pgx.Tx) (res []model.Event, err error)
	FindAllPaginated(ctx context.Context, tx pgx.Tx, param domain.FilterEventParam, pagination domain.PaginationParam) (res entity.PaginatedEvents, err error)
	FindById(ctx context.Context, tx pgx.Tx, eventId string) (event model.Event, err error)
	FindByIdWithVenueAndOrganizer(ctx context.Context, tx pgx.Tx, eventId string) (event entity.Event, err error)
	Count(ctx context.Context, tx pgx.Tx) (res int64, err error)
	Update(ctx context.Context, tx pgx.Tx, event model.Event) (err error)
	SoftDelete(ctx context.Context, tx pgx.Tx, eventId string) (err error)
}

type EventRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewEventRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) EventRepository {
	return &EventRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *EventRepositoryImpl) Create(ctx context.Context, tx pgx.Tx, event model.Event) (id string, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO events (organizer_id, name, description, banner, event_time, status, venue_id, is_active, start_sale_at, end_sale_at, craeted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())`

	if tx != nil {
		_, err = tx.Exec(ctx, query, event.OrganizerID, event.Name, event.Description, event.Banner, event.EventTime, event.Status, event.VenueID, event.IsActive, event.StartSaleAt, event.EndSaleAt)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, event.OrganizerID, event.Name, event.Description, event.Banner, event.EventTime, event.Status, event.VenueID, event.IsActive, event.StartSaleAt, event.EndSaleAt)
	}

	return
}

func (r *EventRepositoryImpl) FindAll(ctx context.Context, tx pgx.Tx) (res []model.Event, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT id, organizer_id, name, description, banner, event_time, status, venue_id, is_active, start_sale_at, end_sale_at, created_at, updated_at FROM events WHERE deleted_at IS NULL`

	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, query)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var event model.Event
		rows.Scan(
			&event.ID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.Banner,
			&event.EventTime,
			&event.Status,
			&event.VenueID,
			&event.IsActive,
			&event.StartSaleAt,
			&event.EndSaleAt,
			&event.CreatedAt,
			&event.UpdatedAt,
		)

		res = append(res, event)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("FindAll event error")
		return res, rows.Err()
	}

	return
}

func (r *EventRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, eventId string) (event model.Event, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT id, organizer_id, name, description, banner, event_time, status, venue_id, additional_information, is_active, start_sale_at, end_sale_at, created_at, updated_at FROM events WHERE id = $1 AND deleted_at IS NULL LIMIT 1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, eventId).Scan(
			&event.ID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.Banner,
			&event.EventTime,
			&event.Status,
			&event.VenueID,
			&event.AdditionalInformation,
			&event.IsActive,
			&event.StartSaleAt,
			&event.EndSaleAt,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, eventId).Scan(
			&event.ID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.Banner,
			&event.EventTime,
			&event.Status,
			&event.VenueID,
			&event.AdditionalInformation,
			&event.IsActive,
			&event.StartSaleAt,
			&event.EndSaleAt,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return event, &lib.ErrorEventNotFound
		}
		return event, err
	}

	return
}

func (r *EventRepositoryImpl) FindByIdWithVenueAndOrganizer(ctx context.Context, tx pgx.Tx, eventId string) (event entity.Event, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT 
		e.id, 
		e.organizer_id, 
		e.name, 
		e.description, 
		e.banner, 
		e.event_time, 
		e.status, 
		e.venue_id, 
		e.additional_information,
		e.is_active, 
		e.start_sale_at, 
		e.end_sale_at, 
		e.created_at, 
		e.updated_at,

		o.name as organizer_name,
		o.slug as organizer_slug,
		o.logo as organizer_logo,

		v.name as venue_name,
		v.venue_type as venue_type,
		v.country as venue_country,
		v.city as venue_city,
		v.capacity as venue_capacity
	FROM events e
		INNER JOIN organizers o ON e.organizer_id = o.id
		INNER JOIN venues v ON e.venue_id = v.id
	WHERE e.id = $1 AND e.deleted_at IS NULL LIMIT 1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, eventId).Scan(
			&event.ID,
			&event.Organizer.ID,
			&event.Name,
			&event.Description,
			&event.Banner,
			&event.EventTime,
			&event.Status,
			&event.Venue.ID,
			&event.AdditionalInformation,
			&event.IsActive,
			&event.StartSaleAt,
			&event.EndSaleAt,
			&event.CreatedAt,
			&event.UpdatedAt,

			&event.Organizer.Name,
			&event.Organizer.Slug,
			&event.Organizer.Logo,

			&event.Venue.Name,
			&event.Venue.VenueType,
			&event.Venue.Country,
			&event.Venue.City,
			&event.Venue.Capacity,
		)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, eventId).Scan(
			&event.ID,
			&event.Organizer.ID,
			&event.Name,
			&event.Description,
			&event.Banner,
			&event.EventTime,
			&event.Status,
			&event.Venue.ID,
			&event.AdditionalInformation,
			&event.IsActive,
			&event.StartSaleAt,
			&event.EndSaleAt,
			&event.CreatedAt,
			&event.UpdatedAt,

			&event.Organizer.Name,
			&event.Organizer.Slug,
			&event.Organizer.Logo,

			&event.Venue.Name,
			&event.Venue.VenueType,
			&event.Venue.Country,
			&event.Venue.City,
			&event.Venue.Capacity,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return event, &lib.ErrorEventNotFound
		}
		return event, err
	}

	return
}

func (r *EventRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, event model.Event) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE events SET
		organizer_id = COALESCE($1, organizer_id), 
		name = COALESCE($2, name), 
		description = COALESCE($3, description), 
		banner = COALESCE($4, banner), 
		event_time = COALESCE($5, event_time), 
		status = COALESCE($6, status), 
		venue_id = COALESCE($7, venue_id), 
		is_active = COALESCE($8, is_active), 
		start_sale_at = COALESCE($9, start_sale_at), 
		end_sale_at = COALESCE($10, end_sale_at), 
		updated_at = CURRENT_TIMESTAMP
		WHERE id = $11 AND deleted_at IS NULL`

	var cmdTag pgconn.CommandTag

	if tx != nil {
		cmdTag, err = tx.Exec(ctx, query,
			event.OrganizerID,
			event.Name,
			event.Description,
			event.Banner,
			event.EventTime,
			event.Status,
			event.VenueID,
			event.IsActive,
			event.StartSaleAt,
			event.EndSaleAt,
			event.ID,
		)
	} else {
		cmdTag, err = r.WrapDB.Postgres.Exec(ctx, query,
			event.OrganizerID,
			event.Name,
			event.Description,
			event.Banner,
			event.EventTime,
			event.Status,
			event.VenueID,
			event.IsActive,
			event.StartSaleAt,
			event.EndSaleAt,
			event.ID,
		)
	}

	// TODO: Check if the update wasn't successful
	if cmdTag.RowsAffected() == 0 {
	}

	return
}

func (r *EventRepositoryImpl) SoftDelete(ctx context.Context, tx pgx.Tx, eventId string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `UPDATE events SET
		deleted_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL`

	// var cmdTag pgCommandTag
	if tx != nil {
		_, err = tx.Exec(ctx, query, eventId)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, eventId)
	}

	return
}

func (r *EventRepositoryImpl) Count(ctx context.Context, tx pgx.Tx) (res int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `SELECT count(id) FROM events WHERE deleted_at IS NULL`

	if tx != nil {
		err = tx.QueryRow(ctx, query).Scan(&res)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query).Scan(&res)
	}

	return
}

func (r *EventRepositoryImpl) FindAllPaginated(ctx context.Context, tx pgx.Tx, param domain.FilterEventParam, pagination domain.PaginationParam) (res entity.PaginatedEvents, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	totalRecords, err := r.Count(ctx, tx)
	if err != nil {
		return
	}

	if totalRecords <= 0 {
		return
	}

	var totalPage int64
	if totalRecords < lib.PaginationPerPage {
		totalPage = 1
	} else {
		totalPage = int64(totalRecords / lib.PaginationPerPage)
		if totalRecords%lib.PaginationPerPage > 0 {
			totalPage += 1
		}
	}

	if pagination.TargetPage < 1 {
		err = &lib.ErrorPaginationPageIsInvalid
		return
	}

	if pagination.TargetPage > totalPage {
		err = &lib.ErrorPaginationReachMaxPage
		return
	}

	additionalParam := ""
	if param.Search != "" {
		additionalParam = fmt.Sprintf("e.name ILIKE '%%%s%%' AND", param.Search)
	}

	if param.Status != "" {
		additionalParam += fmt.Sprintf("e.status = '%s' AND", param.Status)
	}

	query := fmt.Sprintf(`SELECT 
		e.id, 
		e.organizer_id, 
		e.name, 
		e.description, 
		e.banner, 
		e.event_time, 
		e.status, 
		e.venue_id, 
		e.is_active, 
		e.start_sale_at, 
		e.end_sale_at, 
		e.created_at, 
		e.updated_at,

		o.name as organizer_name,
		o.slug as organizer_slug,
		o.logo as organizer_logo,

		v.name as venue_name,
		v.venue_type as venue_type,
		v.country as venue_country,
		v.city as venue_city,
		v.capacity as venue_capacity
	FROM events AS e
		INNER JOIN organizers o ON e.organizer_id = o.id
		INNER JOIN venues v ON e.venue_id = v.id
	WHERE %s e.deleted_at IS NULL 
	ORDER BY $1 %s
	LIMIT $2
	OFFSET $3`, additionalParam, pagination.Order)

	var rows pgx.Rows
	var resPagination entity.Pagination

	resPagination.Page = pagination.TargetPage
	resPagination.TotalRecords = totalRecords
	resPagination.TotalPage = totalPage

	var offsetParam int64 = 0
	if pagination.TargetPage > 1 {
		offsetParam = pagination.TargetPage * lib.PaginationPerPage
	}

	if pagination.TargetPage+1 >= totalPage {
		resPagination.HasNextPage = false
		resPagination.NextPage = pagination.TargetPage
	} else {
		resPagination.HasNextPage = true
		resPagination.NextPage = pagination.TargetPage + 1
	}

	if pagination.TargetPage-1 <= 0 {
		resPagination.HasPreviousPage = false
		resPagination.PreviousPage = pagination.TargetPage
	} else {
		resPagination.HasPreviousPage = true
		resPagination.PreviousPage = pagination.TargetPage - 1
	}

	if tx != nil {
		rows, err = tx.Query(ctx, query, pagination.Order, lib.PaginationPerPage, offsetParam)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, pagination.Order, lib.PaginationPerPage, offsetParam)
	}

	if err != nil {
		return
	}

	defer rows.Close()

	var events []entity.Event = make([]entity.Event, 0)

	for rows.Next() {
		var event entity.Event
		rows.Scan(
			&event.ID,
			&event.Organizer.ID,
			&event.Name,
			&event.Description,
			&event.Banner,
			&event.EventTime,
			&event.Status,
			&event.Venue.ID,
			&event.IsActive,
			&event.StartSaleAt,
			&event.EndSaleAt,
			&event.CreatedAt,
			&event.UpdatedAt,

			&event.Organizer.Name,
			&event.Organizer.Slug,
			&event.Organizer.Logo,

			&event.Venue.Name,
			&event.Venue.VenueType,
			&event.Venue.Country,
			&event.Venue.City,
			&event.Venue.Capacity,
		)

		events = append(events, event)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return res, &lib.ErrorEventNotFound
		}
		return res, err
	}

	res = entity.PaginatedEvents{
		Events:     events,
		Pagination: resPagination,
	}

	return
}
