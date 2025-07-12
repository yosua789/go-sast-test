package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/repository"
)

type Repository struct {
	OrganizerRepo           repository.OrganizerRepository
	VenueRepo               repository.VenueRepository
	VenueSectorRepo         repository.VenueSectorRepository
	EventRepo               repository.EventRepository
	EventSettingRepo        repository.EventSettingsRepository
	EventTicketCategoryRepo repository.EventTicketCategoryRepository
}

func Newrepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) Repository {
	return Repository{
		OrganizerRepo:           repository.NewOrganizerRepository(wrapDB, env),
		VenueRepo:               repository.NewVenueRepository(wrapDB, env),
		VenueSectorRepo:         repository.NewVenueSectorRepository(wrapDB, env),
		EventRepo:               repository.NewEventRepository(wrapDB, env),
		EventSettingRepo:        repository.NewEventSettingsRepository(wrapDB, env),
		EventTicketCategoryRepo: repository.NewEventTicketCategoryRepository(wrapDB, env),
	}
}
