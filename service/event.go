package service

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/domain"
	"assist-tix/dto"
	"assist-tix/helper"
	"assist-tix/lib"
	"assist-tix/model"
	"assist-tix/repository"
	"context"
)

type EventService interface {
	CreateEvent(ctx context.Context, req dto.CreateEventRequest) (res dto.EventResponse, err error)
	GetAllEvent(ctx context.Context) (res []dto.EventResponse, err error)
	GetAllEventPaginated(ctx context.Context, filter dto.FilterEventRequest, pagination dto.PaginationParam) (res dto.PaginatedEvents, err error)
	GetEventById(ctx context.Context, eventId string) (res dto.DetailEventResponse, err error)
	Update(ctx context.Context, eventId string, req dto.EventResponse) (err error)
	Delete(ctx context.Context, eventId string) (err error)
}

type EventServiceImpl struct {
	DB                      *database.WrapDB
	Env                     *config.EnvironmentVariable
	EventRepo               repository.EventRepository
	EventSettingRepo        repository.EventSettingsRepository
	EventTicketCategoryRepo repository.EventTicketCategoryRepository
	OrganizerRepo           repository.OrganizerRepository
	VenueRepo               repository.VenueRepository
}

func NewEventService(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	eventRepo repository.EventRepository,
	eventSettingRepo repository.EventSettingsRepository,
	eventTicketCategoryRepo repository.EventTicketCategoryRepository,
	organizerRepo repository.OrganizerRepository,
	venueRepo repository.VenueRepository,
) EventService {
	return &EventServiceImpl{
		DB:                      db,
		Env:                     env,
		EventRepo:               eventRepo,
		EventSettingRepo:        eventSettingRepo,
		EventTicketCategoryRepo: eventTicketCategoryRepo,
		OrganizerRepo:           organizerRepo,
		VenueRepo:               venueRepo,
	}
}

// TODO
func (s *EventServiceImpl) CreateEvent(ctx context.Context, req dto.CreateEventRequest) (res dto.EventResponse, err error) {
	return
}

func (s *EventServiceImpl) GetAllEvent(ctx context.Context) (res []dto.EventResponse, err error) {
	events, err := s.EventRepo.FindAll(ctx, nil)
	if err != nil {
		return
	}

	res = make([]dto.EventResponse, 0)

	if len(events) < 1 {
		return
	}

	var organizerIds []string = make([]string, 0)
	for _, val := range events {
		organizerIds = append(organizerIds, val.OrganizerID)
	}

	var venueIds []string = make([]string, 0)
	for _, val := range events {
		venueIds = append(venueIds, val.VenueID)
	}

	var organizerMap map[string]model.Organizer = make(map[string]model.Organizer)
	organizers, err := s.OrganizerRepo.FindByIds(ctx, nil, organizerIds...)
	if err != nil {
		return
	}

	for _, val := range organizers {
		_, ok := organizerMap[val.ID]
		if !ok {
			organizerMap[val.ID] = val
		}
	}

	var venueMap map[string]model.Venue = make(map[string]model.Venue)
	venues, err := s.VenueRepo.FindByIds(ctx, nil, venueIds...)
	if err != nil {
		return
	}

	for _, val := range venues {
		_, ok := venueMap[val.ID]
		if !ok {
			venueMap[val.ID] = val
		}
	}

	res = make([]dto.EventResponse, 0)

	for _, val := range events {
		organizer, ok := organizerMap[val.OrganizerID]
		if !ok {
			organizer = model.Organizer{}
		}

		venue, ok := venueMap[val.VenueID]
		if !ok {
			venue = model.Venue{}
		}

		res = append(res, dto.EventResponse{
			ID:          val.ID,
			Organizer:   lib.MapOrganizerModelToSimpleResponse(organizer),
			Name:        val.Name,
			Description: val.Description,
			Banner:      val.Banner,
			EventTime:   val.EventTime,
			Status:      val.Status,
			Venue:       lib.MapVenueModelToSimpleResponse(venue),

			StartSaleAt: helper.ConvertNullTimeToPointer(val.StartSaleAt),
			EndSaleAt:   helper.ConvertNullTimeToPointer(val.EndSaleAt),

			CreatedAt: val.CreatedAt,
			UpdatedAt: helper.ConvertNullTimeToPointer(val.UpdatedAt),
		})
	}

	return
}

func (s *EventServiceImpl) GetEventById(ctx context.Context, eventId string) (res dto.DetailEventResponse, err error) {
	event, err := s.EventRepo.FindByIdWithVenueAndOrganizer(ctx, nil, eventId)
	if err != nil {
		return
	}

	eventSettings, err := s.EventSettingRepo.FindByEventId(ctx, nil, eventId)
	if err != nil {
		return
	}

	eventResponse := lib.MapEventSettingEntityToEventSettingResponse(eventSettings)

	ticketCategories, err := s.EventTicketCategoryRepo.FindByEventId(ctx, nil, eventId)
	if err != nil {
		return
	}

	var ticketCategoriesResponse []dto.EventTicketCategoryResponse = make([]dto.EventTicketCategoryResponse, 0)
	for _, ticketCategory := range ticketCategories {
		ticketCategoryResponse := lib.MapEventTicketCategoryModelToEventTicketCategoryResponse(ticketCategory)
		ticketCategoriesResponse = append(ticketCategoriesResponse, ticketCategoryResponse)
	}

	res = dto.DetailEventResponse{
		ID:          event.ID,
		Organizer:   lib.MapOrganizerEntityToSimpleResponse(event.Organizer),
		Name:        event.Name,
		Description: event.Description,
		Banner:      event.Banner,
		EventTime:   event.EventTime,
		Status:      event.Status,
		Venue:       lib.MapVenueEntityToSimpleResponse(event.Venue),

		AdditionalInformation: event.AdditionalInformation,

		ActiveSettings: eventResponse,

		TicketCategories: ticketCategoriesResponse,

		StartSaleAt: helper.ConvertNullTimeToPointer(event.StartSaleAt),
		EndSaleAt:   helper.ConvertNullTimeToPointer(event.EndSaleAt),

		CreatedAt: event.CreatedAt,
		UpdatedAt: helper.ConvertNullTimeToPointer(event.UpdatedAt),
	}

	return
}

// TODO
func (s *EventServiceImpl) Update(ctx context.Context, eventId string, req dto.EventResponse) (err error) {
	return
}

func (s *EventServiceImpl) Delete(ctx context.Context, eventId string) (err error) {
	_, err = s.EventRepo.FindById(ctx, nil, eventId)
	if err != nil {
		return
	}

	err = s.EventRepo.SoftDelete(ctx, nil, eventId)
	if err != nil {
		return
	}

	return
}

func (s *EventServiceImpl) GetAllEventPaginated(ctx context.Context, filter dto.FilterEventRequest, pagination dto.PaginationParam) (res dto.PaginatedEvents, err error) {

	filterDB := domain.FilterEventParam{
		Search: filter.Search,
		Status: filter.Status,
	}

	if pagination.TargetPage < 1 {
		pagination.TargetPage = 1
	}

	paginationDB := domain.PaginationParam{
		TargetPage: pagination.TargetPage,
	}

	paginatedEvents, err := s.EventRepo.FindAllPaginated(ctx, nil, filterDB, paginationDB)
	if err != nil {
		return
	}

	res.Events = make([]dto.EventResponse, 0)

	for _, val := range paginatedEvents.Events {
		event := lib.MapEventEntityToEventResponse(val)

		res.Events = append(res.Events, event)
	}

	var prevPage *int64
	if !paginatedEvents.Pagination.HasPreviousPage {
		prevPage = nil
	} else {
		prevPage = &paginatedEvents.Pagination.PreviousPage
	}

	var nextPage *int64
	if !paginatedEvents.Pagination.HasNextPage {
		nextPage = nil
	} else {
		nextPage = &paginatedEvents.Pagination.NextPage
	}

	res = dto.PaginatedEvents{
		Events: res.Events,
		Pagination: dto.Pagination{
			TotalRecords: paginatedEvents.Pagination.TotalRecords,
			MaxPage:      paginatedEvents.Pagination.TotalPage,
			CurrentPage:  paginatedEvents.Pagination.Page,
			PrevPage:     prevPage,
			NextPage:     nextPage,
		},
	}

	return
}
