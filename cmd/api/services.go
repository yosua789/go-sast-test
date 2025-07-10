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
}

func Newservice(
	env *config.EnvironmentVariable,
	r Repository,
	db *database.WrapDB,
) Service {
	organizerService := service.NewOrganizerService(db, env, r.OrganizerRepo)
	venueService := service.NewVenueService(db, env, r.VenueRepo)
	eventService := service.NewEventService(db, env, r.EventRepo, r.EventSettingRepo, r.EventTicketCategoryRepo, r.OrganizerRepo, r.VenueRepo)
	eventTicketCategoryService := service.NewEventTicketCategoryService(db, env, r.VenueRepo, r.EventRepo, r.EventTicketCategoryRepo)
	return Service{
		OrganizerService:           organizerService,
		VenueService:               venueService,
		EventService:               eventService,
		EventTicketCategoryService: eventTicketCategoryService,
	}
}
