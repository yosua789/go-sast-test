package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/service"
)

type Service struct {
	OrganizerService           service.OrganizerService
	VenueService               service.VenueService
	EventService               service.EventService
	EventTicketCategoryService service.EventTicketCategoryService
	EventTransactionService    service.EventTransactionService
}

func Newservice(
	env *config.EnvironmentVariable,
	r Repository,
	db *database.WrapDB,
	job Job,
) Service {
	organizerService := service.NewOrganizerService(db, env, r.OrganizerRepo)
	venueService := service.NewVenueService(db, env, r.VenueRepo, r.VenueSectorRepo)
	eventService := service.NewEventService(db, env, r.EventRepo, r.EventSettingRepo, r.EventTicketCategoryRepo, r.OrganizerRepo, r.VenueRepo, r.EventTransactionGarudaIDRepo, r.GcsStorageRepository)
	eventTicketCategoryService := service.NewEventTicketCategoryService(db, env, r.VenueRepo, r.VenueSectorRepo, r.EventRepo, r.EventTicketCategoryRepo, r.EventSeatmapBookRepo, r.GcsStorageRepository)
	eventTransactionService := service.NewEventTransactionService(db, env, r.EventRepo, r.EventSettingRepo, r.EventTicketCategoryRepo, r.EventTransactionRepo, r.EventTransactionItemRepo, r.EventSeatmapBookRepo, r.VenueSectorRepo, r.EventTransactionGarudaIDRepo, job.CheckStatusTransactionJob)

	return Service{
		OrganizerService:           organizerService,
		VenueService:               venueService,
		EventService:               eventService,
		EventTicketCategoryService: eventTicketCategoryService,
		EventTransactionService:    eventTransactionService,
	}
}
