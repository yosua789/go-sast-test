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

	"github.com/rs/zerolog/log"
)

type EventService interface {
	CreateEvent(ctx context.Context, req dto.CreateEventRequest) (res dto.EventResponse, err error)
	GetAllEvent(ctx context.Context) (res []dto.EventResponse, err error)
	GetAllEventPaginated(ctx context.Context, filter dto.FilterEventRequest, pagination dto.PaginationParam) (res dto.PaginatedEvents, err error)
	GetEventById(ctx context.Context, eventId string) (res dto.DetailEventResponse, err error)
	Update(ctx context.Context, eventId string, req dto.EventResponse) (err error)
	Delete(ctx context.Context, eventId string) (err error)
	FindByGarudaID(ctx context.Context, eventID, garudaID string) (dto.VerifyGarudaIDResponse, error)
}

type EventServiceImpl struct {
	DB                           *database.WrapDB
	Env                          *config.EnvironmentVariable
	EventRepo                    repository.EventRepository
	EventSettingRepo             repository.EventSettingsRepository
	EventTicketCategoryRepo      repository.EventTicketCategoryRepository
	OrganizerRepo                repository.OrganizerRepository
	VenueRepo                    repository.VenueRepository
	VenueSectorRepo              repository.VenueSectorRepository
	EventTransactionGarudaIDRepo repository.EventTransactionGarudaIDRepository

	GCSStorageRepo repository.GCSStorageRepository
}

func NewEventService(
	db *database.WrapDB,
	env *config.EnvironmentVariable,
	eventRepo repository.EventRepository,
	eventSettingRepo repository.EventSettingsRepository,
	eventTicketCategoryRepo repository.EventTicketCategoryRepository,
	organizerRepo repository.OrganizerRepository,
	venueRepo repository.VenueRepository,
	eventTransactionGarudaIDRepo repository.EventTransactionGarudaIDRepository,
	gcsStorageRepo repository.GCSStorageRepository,
) EventService {
	return &EventServiceImpl{
		DB:                           db,
		Env:                          env,
		EventRepo:                    eventRepo,
		EventSettingRepo:             eventSettingRepo,
		EventTicketCategoryRepo:      eventTicketCategoryRepo,
		OrganizerRepo:                organizerRepo,
		VenueRepo:                    venueRepo,
		EventTransactionGarudaIDRepo: eventTransactionGarudaIDRepo,
		GCSStorageRepo:               gcsStorageRepo,
	}
}

// TODO
func (s *EventServiceImpl) CreateEvent(ctx context.Context, req dto.CreateEventRequest) (res dto.EventResponse, err error) {
	return
}

func (s *EventServiceImpl) GetAllEvent(ctx context.Context) (res []dto.EventResponse, err error) {
	log.Info().Msg("Get all events")
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
	log.Info().Any("OrganizerIds", organizerIds).Msg("Mapping organizer ids")

	var venueIds []string = make([]string, 0)
	for _, val := range events {
		venueIds = append(venueIds, val.VenueID)
	}
	log.Info().Any("VenueIds", venueIds).Msg("Mapping venue ids")

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

	log.Info().Msg("Mapping events with venue & organizer")
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
			Venue:       lib.MapVenueModelToSimpleResponse(venue),

			StartSaleAt: helper.ConvertNullTimeToPointer(val.StartSaleAt),
			EndSaleAt:   helper.ConvertNullTimeToPointer(val.EndSaleAt),

			CreatedAt: val.CreatedAt,
			UpdatedAt: helper.ConvertNullTimeToPointer(val.UpdatedAt),
		})
	}

	log.Info().Int("Count", len(events)).Msg("Events success")

	return
}

func (s *EventServiceImpl) GetEventById(ctx context.Context, eventId string) (res dto.DetailEventResponse, err error) {
	log.Info().Str("EventID", eventId).Msg("Get event by ID")

	event, err := s.EventRepo.FindByIdWithVenueAndOrganizer(ctx, nil, eventId)
	if err != nil {
		return
	}

	if s.Env.Storage.Type == lib.StorageTypeGCS {
		bannerUrl, errBanner := s.GCSStorageRepo.CreateSignedUrl(event.Banner)
		if errBanner != nil {
			err = errBanner
			return
		}

		event.Banner = bannerUrl
	}

	log.Info().Msg("Get event settings by event id")
	eventSettings, err := s.EventSettingRepo.FindByEventId(ctx, nil, eventId)
	if err != nil {
		return
	}
	log.Info().Interface("SettingsRaw", eventSettings).Msg("mapping event settings")

	eventSettingsResponse := lib.MapEventSettingEntityToEventSettingResponse(eventSettings)
	log.Info().Interface("SettingsResponse", eventSettingsResponse).Msg("Event settings")

	log.Info().Msg("Get ticket categories by event id")
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
		ID:           event.ID,
		Organizer:    lib.MapOrganizerEntityToSimpleResponse(event.Organizer),
		Name:         event.Name,
		Description:  event.Description,
		Banner:       event.Banner,
		EventTime:    event.EventTime,
		Venue:        lib.MapVenueEntityToSimpleResponse(event.Venue),
		IsSaleActive: event.IsSaleActive,

		AdditionalInformation: event.AdditionalInformation,

		ActiveSettings: eventSettingsResponse,

		TicketCategories: ticketCategoriesResponse,

		StartSaleAt: helper.ConvertNullTimeToPointer(event.StartSaleAt),
		EndSaleAt:   helper.ConvertNullTimeToPointer(event.EndSaleAt),

		CreatedAt: event.CreatedAt,
		UpdatedAt: helper.ConvertNullTimeToPointer(event.UpdatedAt),
	}

	log.Info().Int("TicketCategory", len(ticketCategoriesResponse)).Msg("Get event by id success")

	return
}

// TODO
func (s *EventServiceImpl) Update(ctx context.Context, eventId string, req dto.EventResponse) (err error) {
	return
}

func (s *EventServiceImpl) Delete(ctx context.Context, eventId string) (err error) {
	log.Info().Str("eventId", eventId).Msg("Delete event by id")
	_, err = s.EventRepo.FindById(ctx, nil, eventId)
	if err != nil {
		return
	}

	err = s.EventRepo.SoftDelete(ctx, nil, eventId)
	if err != nil {
		return
	}

	log.Info().Msg("Success delete event")

	return
}

func (s *EventServiceImpl) GetAllEventPaginated(ctx context.Context, filter dto.FilterEventRequest, pagination dto.PaginationParam) (res dto.PaginatedEvents, err error) {
	log.Info().Str("Search", filter.Search).Str("Status", filter.Status).Int("TargetPage", int(pagination.TargetPage)).Msg("Get paginated events")

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

		if s.Env.Storage.Type == lib.StorageTypeGCS {
			signedUrl, err := s.GCSStorageRepo.CreateSignedUrl(event.Banner)
			if err != nil {
				continue
			}
			event.Banner = signedUrl
		}

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

	log.Info().Int("totalRecords", int(paginatedEvents.Pagination.TotalRecords)).Int("MaxPage", int(paginatedEvents.Pagination.TotalPage)).Msg("Success get paginated events")

	return
}

func (s *EventServiceImpl) FindByGarudaID(ctx context.Context, garudaID, eventID string) (resp dto.VerifyGarudaIDResponse, err error) {

	ctx, cancel := context.WithTimeout(ctx, s.Env.Database.Timeout.Write)
	defer cancel()

	_, err = s.EventRepo.FindById(ctx, nil, eventID)
	if err != nil {
		log.Error().Err(err).Msg("failed to find event by id")
		// return resp, &lib.ErrorEventNotFound if event not found
		return
	}

	settings, err := s.EventSettingRepo.FindByEventId(ctx, nil, eventID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get event settings")
		return
	}
	log.Info().Interface("SettingsRaw", settings).Msg("mapping event settings")
	eventSettings := lib.MapEventSettings(settings)
	log.Info().Interface("Settings", eventSettings).Msg("Event settings")
	if !eventSettings.GarudaIdVerification {
		log.Info().Msg("Garuda ID verification is not enabled for this event")
		return dto.VerifyGarudaIDResponse{IsAvailable: false}, &lib.ErrorEventNonGarudaID
	}
	err = s.EventTransactionGarudaIDRepo.GetEventGarudaID(ctx, eventID, garudaID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get event garuda id")
		if err == &lib.ErrorGarudaIDAlreadyUsed {
			return resp, &lib.ErrorGarudaIDAlreadyUsed
		}
		return resp, &lib.ErrorInternalServer
	}

	externalResp, err := helper.VerifyUserGarudaIDByID(s.Env.GarudaID.BaseUrl, garudaID)
	if err != nil {
		return resp, &lib.ErrorGetGarudaID
	}

	if externalResp != nil && !externalResp.Success {
		log.Info().Int("ErrorCode", externalResp.ErrorCode).Msg("Garuda ID verification failed")
		switch externalResp.ErrorCode {
		case 40401:
			return resp, &lib.ErrorGarudaIDNotFound
		case 42205:
			return resp, &lib.ErrorGarudaIDBlacklisted
		case 40909:
			return resp, &lib.ErrorGarudaIDInvalid
		case 40910:
			return resp, &lib.ErrorGarudaIDRejected
		case 50001:
			return resp, &lib.ErrorGetGarudaID
		}
	}
	resp.IsAvailable = true
	resp.GarudaID = garudaID
	return resp, nil
}
