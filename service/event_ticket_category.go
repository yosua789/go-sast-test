package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/dto"
	"assist-tix/helper"
	"assist-tix/lib"
	"assist-tix/model"
	"assist-tix/repository"
	"context"

	"github.com/rs/zerolog/log"
)

type EventTicketCategoryService interface {
	Create(ctx context.Context, eventId string, req dto.CreateEventTicketCategoryRequest) (err error)
	GetVenueTicketsByEventId(ctx context.Context, eventId string) (res dto.VenueEventTicketCategoryResponse, err error)
	GetByEventId(ctx context.Context, eventId string) (res []dto.DetailEventTicketCategoryResponse, err error)
	GetById(ctx context.Context, eventId string, ticketCategoryId string) (res dto.DetailEventTicketCategoryResponse, err error)
	GetSeatmapByTicketCategoryId(ctx context.Context, eventId, ticketCategoryId string) (res dto.EventSectorSeatmapResponse, err error)
	Delete(ctx context.Context, eventId, ticketCategoryId string) (err error)
}

type EventTicketCategoryServiceImpl struct {
	DB                            *database.WrapDB
	Env                           *config.EnvironmentVariable
	VenueRepository               repository.VenueRepository
	VenueSectorRepository         repository.VenueSectorRepository
	EventRepository               repository.EventRepository
	EventTicketCategoryRepository repository.EventTicketCategoryRepository
	EventSeatmapBookRepository    repository.EventSeatmapBookRepository

	GCSStorageRepo repository.GCSStorageRepository
}

func NewEventTicketCategoryService(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	venueRepository repository.VenueRepository,
	venueSectorRepository repository.VenueSectorRepository,
	eventRepository repository.EventRepository,
	eventTicketCategoryRepository repository.EventTicketCategoryRepository,
	eventSeatmapBookRepository repository.EventSeatmapBookRepository,
	gcsStorageRepo repository.GCSStorageRepository,
) EventTicketCategoryService {
	return &EventTicketCategoryServiceImpl{
		DB:                            db,
		Env:                           env,
		VenueRepository:               venueRepository,
		VenueSectorRepository:         venueSectorRepository,
		EventRepository:               eventRepository,
		EventTicketCategoryRepository: eventTicketCategoryRepository,
		EventSeatmapBookRepository:    eventSeatmapBookRepository,
		GCSStorageRepo:                gcsStorageRepo,
	}
}

func (s *EventTicketCategoryServiceImpl) Create(ctx context.Context, eventId string, req dto.CreateEventTicketCategoryRequest) (err error) {

	log.Info().Str("eventId", eventId).Str("name", req.Name).Msg("create event ticket category")
	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Validate event id
	log.Info().Msg("validate and find event by id")
	event, err := s.EventRepository.FindById(ctx, tx, eventId)
	if err != nil {
		return
	}

	// Validate Sector is same venue with event
	log.Info().Str("sectorId", req.SectorID).Msg("validate and find venue sector by id")
	sector, err := s.VenueSectorRepository.FindById(ctx, tx, req.SectorID)
	if err != nil {
		return
	}

	log.Info().Str("sectorId", req.SectorID).Str("venueId", event.VenueID).Msg("validate sector venue is same with event venue")
	if sector.VenueID != event.VenueID {
		err = &lib.ErrorVenueSectorNotFound
		return
	}

	createEventTicketCategory := model.EventTicketCategory{
		EventID:              eventId,
		VenueSectorId:        req.SectorID,
		Name:                 req.Name,
		Description:          req.Description,
		Price:                req.Price,
		TotalStock:           req.TotalStock,
		TotalPublicStock:     req.TotalPublicStock,
		PublicStock:          req.PublicStock,
		TotalComplimentStock: req.TotalComplimentStock,
		ComplimentStock:      req.ComplimentStock,
		Code:                 req.Code,
		Entrance:             req.Entrance,
	}

	log.Info().Msg("insert data event ticket category")
	err = s.EventTicketCategoryRepository.Create(ctx, nil, createEventTicketCategory)
	if err != nil {
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	log.Info().Msg("success create event ticket category")

	return
}

func (s *EventTicketCategoryServiceImpl) GetByEventId(ctx context.Context, eventId string) (res []dto.DetailEventTicketCategoryResponse, err error) {
	// Validate event id
	log.Info().Str("eventId", eventId).Msg("validate event")
	_, err = s.EventRepository.FindById(ctx, nil, eventId)
	if err != nil {
		return
	}

	log.Info().Str("eventId", eventId).Msg("get ticket categories by event id")
	ticketCategories, err := s.EventTicketCategoryRepository.FindByEventId(ctx, nil, eventId)
	if err != nil {
		return
	}

	res = make([]dto.DetailEventTicketCategoryResponse, 0)
	for _, val := range ticketCategories {
		res = append(res, lib.MapEventTicketCategoryModelToDetailEventTicketCategoryResponse(val))
	}

	log.Info().Int("count", len(res)).Msg("success get ticket categories by event id")

	return
}

func (s *EventTicketCategoryServiceImpl) GetVenueTicketsByEventId(ctx context.Context, eventId string) (res dto.VenueEventTicketCategoryResponse, err error) {
	// Validate event id
	log.Info().Str("eventId", eventId).Msg("validate event")
	event, err := s.EventRepository.FindById(ctx, nil, eventId)
	if err != nil {
		return
	}

	log.Info().Str("venueId", event.VenueID).Msg("find venue")
	venue, err := s.VenueRepository.FindById(ctx, nil, event.VenueID)
	if err != nil {
		return
	}

	if s.Env.Storage.Type == lib.StorageTypeGCS {
		signedUrl, errSignedUrl := s.GCSStorageRepo.CreateSignedUrl(venue.Image)
		if errSignedUrl != nil {
			err = errSignedUrl
			return
		}

		venue.Image = signedUrl
	}

	log.Info().Str("eventId", eventId).Msg("find ticket categories by event id")
	ticketCategories, err := s.EventTicketCategoryRepository.FindTicketSectorsByEventId(ctx, nil, eventId)
	if err != nil {
		return
	}

	tickets := make([]dto.DetailEventPublicTicketCategoryResponse, 0)
	for _, val := range ticketCategories {
		if val.TotalPublicStock == 0 && val.PublicStock == 0 {
			tickets = append(tickets, lib.MapEntityTicketCategoryToDetailEventPublicTicketCategoryResponse(val))
		}
	}

	log.Info().Int("count", len(tickets)).Msg("tickets")

	res = dto.VenueEventTicketCategoryResponse{
		Venue:            lib.MapVenueModelToVenueResponse(venue),
		TicketCategories: tickets,
	}

	log.Info().Msg("success get venue tickets by event id")

	return
}

func (s *EventTicketCategoryServiceImpl) GetById(ctx context.Context, eventId string, ticketCategoryId string) (res dto.DetailEventTicketCategoryResponse, err error) {
	log.Info().Str("eventId", eventId).Str("ticketCategoryId", ticketCategoryId).Msg("get ticket category by id")
	ticketCategory, err := s.EventTicketCategoryRepository.FindByIdAndEventId(ctx, nil, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	res = lib.MapEventTicketCategoryModelToDetailEventTicketCategoryResponse(ticketCategory)
	log.Info().Msg("success get ticket category by id")
	return
}

func (s *EventTicketCategoryServiceImpl) Delete(ctx context.Context, eventId, ticketCategoryId string) (err error) {
	log.Info().Str("eventId", eventId).Str("ticketCategoryId", ticketCategoryId).Msg("validate ticket category")
	_, err = s.EventTicketCategoryRepository.FindByIdAndEventId(ctx, nil, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	log.Info().Msg("delete ticket category from database")
	err = s.EventTicketCategoryRepository.SoftDelete(ctx, nil, ticketCategoryId)
	if err != nil {
		return
	}

	log.Info().Msg("success delete ticket category by id")

	return
}

func (s *EventTicketCategoryServiceImpl) GetSeatmapByTicketCategoryId(ctx context.Context, eventId, ticketCategoryId string) (res dto.EventSectorSeatmapResponse, err error) {
	log.Info().Msg("get seatmap by ticket category id")

	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	log.Info().Str("eventId", eventId).Str("ticketCategoryId", ticketCategoryId).Msg("find event ticket category")
	eventTickets, err := s.EventTicketCategoryRepository.FindByIdAndEventId(ctx, tx, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	log.Info().Str("venueSectorId", eventTickets.VenueSectorId).Msg("find venue sector")
	sector, err := s.VenueSectorRepository.FindById(ctx, tx, eventTickets.VenueSectorId)
	if err != nil {
		return
	}

	log.Info().Bool("hasSeatmap", sector.HasSeatmap).Msg("check sector has seatmap")
	if !sector.HasSeatmap {
		err = &lib.ErrorVenueSectorDoesntHaveSeatmap
		return
	}

	log.Info().Str("eventId", eventId).Str("sectorId", sector.ID).Msg("find seatmap by event sector id")
	seatmapRes, err := s.EventTicketCategoryRepository.FindSeatmapByEventSectorId(ctx, tx, eventId, sector.ID)
	if err != nil {
		return
	}

	log.Info().Str("eventId", eventId).Str("sectorId", sector.ID).Msg("find seatmap book by event sector id")
	eventSeatmapBooks, err := s.EventSeatmapBookRepository.FindSeatBooksByEventSectorId(ctx, tx, eventId, sector.ID)
	if err != nil {
		return
	}

	log.Info().Msg("mapping seatmap row and column sector")
	var (
		currentRow   int = -1
		currentSeats []dto.SectorSeatmapResponse

		seatmap = make([]dto.SectorSeatmapRowResponse, 0)
	)

	res = dto.EventSectorSeatmapResponse{
		ID:       sector.ID,
		Name:     sector.Name,
		Color:    sector.SectorColor,
		AreaCode: sector.AreaCode,
	}

	for i, val := range seatmapRes {
		if currentRow == -1 {
			currentRow = val.SeatRow
		}

		_, ok := eventSeatmapBooks[helper.ConvertRowColumnKey(val.SeatRow, val.SeatColumn)]
		if ok {
			val.Status = lib.SeatmapStatusUnavailable
		}

		seat := dto.SectorSeatmapResponse{
			Column: val.SeatColumn,
			Label:  val.Label,
			Status: val.Status,
		}

		if val.SeatRow != currentRow {
			seatmap = append(seatmap, dto.SectorSeatmapRowResponse{
				Row:   currentRow,
				Seats: currentSeats,
			})

			currentSeats = nil
			currentRow = val.SeatRow
		}
		currentSeats = append(currentSeats, seat)

		if i == len(seatmapRes)-1 {
			seatmap = append(seatmap, dto.SectorSeatmapRowResponse{
				Row:   currentRow,
				Seats: currentSeats,
			})
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return
	}

	res.Seatmap = seatmap

	log.Info().Msg("success get seatmap by ticket category id")

	return
}
