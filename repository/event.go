package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/domain"
	"assist-tix/entity"
	"assist-tix/helper"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

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
	Count(ctx context.Context, tx pgx.Tx, param *domain.FilterEventParam) (res int64, err error)
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

	query := `INSERT INTO events (organizer_id, name, description, banner_filename, event_time, venue_id, start_sale_at, end_sale_at, craeted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`

	if tx != nil {
		_, err = tx.Exec(ctx, query, event.OrganizerID, event.Name, event.Description, event.Banner, event.EventTime, event.VenueID, event.StartSaleAt, event.EndSaleAt)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, event.OrganizerID, event.Name, event.Description, event.Banner, event.EventTime, event.VenueID, event.StartSaleAt, event.EndSaleAt)
	}

	return
}

func (r *EventRepositoryImpl) FindAll(ctx context.Context, tx pgx.Tx) (res []model.Event, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	query := `SELECT id, organizer_id, name, description, banner_filename, event_time, venue_id, publish_status, is_sale_active, start_sale_at, end_sale_at, created_at, updated_at FROM events WHERE deleted_at IS NULL`

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
			&event.VenueID,
			&event.PublishStatus,
			&event.IsSaleActive,
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

	query := fmt.Sprintf(`SELECT 
		id, 
		organizer_id, 
		name, 
		description, 
		banner_filename, 
		event_time, 
		venue_id, 
		is_sale_active, 
		publish_status, 
		additional_information, 
		start_sale_at, 
		end_sale_at, 
		created_at, 
		updated_at 
	FROM events 
	WHERE id = $1 
		AND (publish_status = '%s' OR publish_status = '%s') 
		AND deleted_at IS NULL
	LIMIT 1`, lib.EventPublishStatusPublished, lib.EventPublishStatusPaused)

	if tx != nil {
		err = tx.QueryRow(ctx, query, eventId).Scan(
			&event.ID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.Banner,
			&event.EventTime,
			&event.VenueID,
			&event.IsSaleActive,
			&event.PublishStatus,
			&event.AdditionalInformation,
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
			&event.VenueID,
			&event.IsSaleActive,
			&event.PublishStatus,
			&event.AdditionalInformation,
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

	query := fmt.Sprintf(`SELECT 
		e.id, 
		e.organizer_id, 
		e.name, 
		e.description, 
		e.banner_filename, 
		e.event_time, 
		e.is_sale_active,
		e.publish_status, 
		e.venue_id, 
		e.additional_information,
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
	WHERE 
		e.id = $1 
		AND (e.publish_status = '%s' OR e.publish_status = '%s')
		AND e.deleted_at IS NULL LIMIT 1`, lib.EventPublishStatusPublished, lib.EventPublishStatusPaused)

	if tx != nil {
		err = tx.QueryRow(ctx, query, eventId).Scan(
			&event.ID,
			&event.Organizer.ID,
			&event.Name,
			&event.Description,
			&event.Banner,
			&event.EventTime,
			&event.IsSaleActive,
			&event.PublishStatus,
			&event.Venue.ID,
			&event.AdditionalInformation,
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
			&event.IsSaleActive,
			&event.PublishStatus,
			&event.Venue.ID,
			&event.AdditionalInformation,
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
		banner_filename = COALESCE($4, banner_filename), 
		event_time = COALESCE($5, event_time), 
		publish_status = COALESCE($6, publish_status), 
		is_sale_active = COALESCE($7, is_sale_active),
		venue_id = COALESCE($8, venue_id), 		
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
			event.PublishStatus,
			event.IsSaleActive,
			event.VenueID,
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
			event.PublishStatus,
			event.IsSaleActive,
			event.VenueID,
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

func (r *EventRepositoryImpl) Count(ctx context.Context, tx pgx.Tx, param *domain.FilterEventParam) (res int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	var (
		args       []interface{}
		conditions []string
		argIndex   = 1
	)

	if param.Search != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+param.Search+"%")
		argIndex++
	}

	switch param.Status {
	case lib.EventStatusUpComing:
		conditions = append(conditions, fmt.Sprintf("event_time >= $%d", argIndex))
		args = append(args, time.Now())
		argIndex++
	case lib.EventStatusFinished:
		conditions = append(conditions, fmt.Sprintf("event_time <= $%d", argIndex))
		args = append(args, time.Now())
		argIndex++
	}

	conditions = append(conditions, fmt.Sprintf("( publish_status = '%s' OR publish_status = '%s' )", lib.EventPublishStatusPublished, lib.EventPublishStatusPaused))
	conditions = append(conditions, "deleted_at IS NULL")

	whereClause := "WHERE " + helper.JoinWithAnd(conditions)

	query := fmt.Sprintf(`
		SELECT count(id)
		FROM events
		%s
	`, whereClause)

	if tx != nil {
		err = tx.QueryRow(ctx, query, args...).Scan(&res)
	} else {
		err = r.WrapDB.Postgres.QueryRow(ctx, query, args...).Scan(&res)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return
	}

	return
}

func (r *EventRepositoryImpl) FindAllPaginated(ctx context.Context, tx pgx.Tx, param domain.FilterEventParam, pagination domain.PaginationParam) (res entity.PaginatedEvents, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Read)
	defer cancel()

	totalRecords, err := r.Count(ctx, tx, &param)
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

	var (
		args       []interface{}
		conditions []string
		argIndex   = 1
	)

	if param.Search != "" {
		conditions = append(conditions, fmt.Sprintf("e.name ILIKE $%d", argIndex))
		args = append(args, "%"+param.Search+"%")
		argIndex++
	}

	switch param.Status {
	case lib.EventStatusUpComing:
		conditions = append(conditions, fmt.Sprintf("e.event_time >= $%d", argIndex))
		args = append(args, time.Now())
		argIndex++
	case lib.EventStatusFinished:
		conditions = append(conditions, fmt.Sprintf("e.event_time <= $%d", argIndex))
		args = append(args, time.Now())
		argIndex++
	}

	conditions = append(conditions, fmt.Sprintf("( e.publish_status = '%s' OR e.publish_status = '%s' ) ", lib.EventPublishStatusPublished, lib.EventPublishStatusPaused))
	conditions = append(conditions, "e.deleted_at IS NULL")

	whereClause := "WHERE " + helper.JoinWithAnd(conditions)

	allowedOrders := map[string]bool{
		"ASC":  true,
		"DESC": true,
	}
	orderDirection := "ASC"
	if allowedOrders[strings.ToUpper(pagination.Order)] {
		orderDirection = strings.ToUpper(pagination.Order)
	}

	var offsetParam int64 = 0
	if pagination.TargetPage > 1 {
		offsetParam = (pagination.TargetPage - 1) * lib.PaginationPerPage
	}

	args = append(args, lib.PaginationPerPage)
	args = append(args, offsetParam)

	query := fmt.Sprintf(`
		SELECT 
			e.id, 
			e.organizer_id, 
			e.name, 
			e.description, 
			e.banner_filename, 
			e.event_time, 
			e.venue_id,
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
		%s
		ORDER BY e.created_at %s
		LIMIT $%d
		OFFSET $%d
	`, whereClause, orderDirection, argIndex, argIndex+1)

	var rows pgx.Rows
	var resPagination entity.Pagination

	resPagination.Page = pagination.TargetPage
	resPagination.TotalRecords = totalRecords
	resPagination.TotalPage = totalPage

	if pagination.TargetPage >= totalPage {
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
		rows, err = tx.Query(ctx, query, args...)
	} else {
		rows, err = r.WrapDB.Postgres.Query(ctx, query, args...)
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
			&event.Venue.ID,
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
