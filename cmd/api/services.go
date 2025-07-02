package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/service"
)

type Service struct {
	OrganizerService service.OrganizerService
	VenueService     service.VenueService
}

func Newservice(
	env *config.EnvironmentVariable,
	r Repository,
	db *database.WrapDB,
) Service {
	organizerService := service.NewOrganizerService(db, env, r.OrganizerRepo)
	venueService := service.NewVenueService(db, env, r.VenueRepo)
	return Service{
		OrganizerService: organizerService,
		VenueService:     venueService,
	}
}
