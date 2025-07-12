package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/dto"
	"assist-tix/lib"
	"assist-tix/model"
	"assist-tix/repository"
	"context"
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
}

func NewEventTicketCategoryService(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	venueRepository repository.VenueRepository,
	venueSectorRepository repository.VenueSectorRepository,
	eventRepository repository.EventRepository,
	eventTicketCategoryRepository repository.EventTicketCategoryRepository,
) EventTicketCategoryService {
	return &EventTicketCategoryServiceImpl{
		DB:                            db,
		Env:                           env,
		VenueRepository:               venueRepository,
		VenueSectorRepository:         venueSectorRepository,
		EventRepository:               eventRepository,
		EventTicketCategoryRepository: eventTicketCategoryRepository,
	}
}

func (s *EventTicketCategoryServiceImpl) Create(ctx context.Context, eventId string, req dto.CreateEventTicketCategoryRequest) (err error) {

	tx, err := s.DB.Postgres.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Validate event id
	event, err := s.EventRepository.FindById(ctx, tx, eventId)
	if err != nil {
		return
	}

	// Validate Sector is same venue with event
	sector, err := s.VenueSectorRepository.FindById(ctx, tx, req.SectorID)
	if err != nil {
		return
	}

	if sector.VenueID != event.ID {
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

	err = s.EventTicketCategoryRepository.Create(ctx, nil, createEventTicketCategory)
	if err != nil {
		return
	}

	return
}

func (s *EventTicketCategoryServiceImpl) GetByEventId(ctx context.Context, eventId string) (res []dto.DetailEventTicketCategoryResponse, err error) {
	// Validate event id
	_, err = s.EventRepository.FindById(ctx, nil, eventId)
	if err != nil {
		return
	}

	ticketCategories, err := s.EventTicketCategoryRepository.FindByEventId(ctx, nil, eventId)
	if err != nil {
		return
	}

	res = make([]dto.DetailEventTicketCategoryResponse, 0)
	for _, val := range ticketCategories {
		res = append(res, lib.MapEventTicketCategoryModelToDetailEventTicketCategoryResponse(val))
	}

	return
}

func (s *EventTicketCategoryServiceImpl) GetVenueTicketsByEventId(ctx context.Context, eventId string) (res dto.VenueEventTicketCategoryResponse, err error) {
	// Validate event id
	event, err := s.EventRepository.FindById(ctx, nil, eventId)
	if err != nil {
		return
	}

	venue, err := s.VenueRepository.FindById(ctx, nil, event.VenueID)
	if err != nil {
		return
	}

	ticketCategories, err := s.EventTicketCategoryRepository.FindTicketSectorsByEventId(ctx, nil, eventId)
	if err != nil {
		return
	}

	tickets := make([]dto.DetailEventPublicTicketCategoryResponse, 0)
	for _, val := range ticketCategories {
		tickets = append(tickets, lib.MapEntityTicketCategoryToDetailEventPublicTicketCategoryResponse(val))
	}

	res = dto.VenueEventTicketCategoryResponse{
		Venue:            lib.MapVenueModelToVenueResponse(venue),
		TicketCategories: tickets,
	}

	return
}

func (s *EventTicketCategoryServiceImpl) GetById(ctx context.Context, eventId string, ticketCategoryId string) (res dto.DetailEventTicketCategoryResponse, err error) {
	ticketCategory, err := s.EventTicketCategoryRepository.FindByIdAndEventId(ctx, nil, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	res = lib.MapEventTicketCategoryModelToDetailEventTicketCategoryResponse(ticketCategory)
	return
}

func (s *EventTicketCategoryServiceImpl) Delete(ctx context.Context, eventId, ticketCategoryId string) (err error) {
	_, err = s.EventTicketCategoryRepository.FindByIdAndEventId(ctx, nil, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	err = s.EventTicketCategoryRepository.SoftDelete(ctx, nil, ticketCategoryId)
	if err != nil {
		return
	}

	return
}

// TODO: Implement Booked seat, mark seat as unavailable when seat is booked
func (s *EventTicketCategoryServiceImpl) GetSeatmapByTicketCategoryId(ctx context.Context, eventId, ticketCategoryId string) (res dto.EventSectorSeatmapResponse, err error) {
	eventTickets, err := s.EventTicketCategoryRepository.FindByIdAndEventId(ctx, nil, eventId, ticketCategoryId)
	if err != nil {
		return
	}

	sector, err := s.VenueSectorRepository.FindById(ctx, nil, eventTickets.VenueSectorId)
	if err != nil {
		return
	}

	if !sector.HasSeatmap {
		err = &lib.ErrorVenueSectorDoesntHaveSeatmap
		return
	}

	seatmapRes, err := s.EventTicketCategoryRepository.FindSeatmapByEventSectorId(ctx, nil, eventId, sector.ID)
	if err != nil {
		return
	}

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

	res.Seatmap = seatmap

	return
}
