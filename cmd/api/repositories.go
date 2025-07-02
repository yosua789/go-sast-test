package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/repository"
)

type Repository struct {
	OrganizerRepo repository.OrganizerRepository
	VenueRepo     repository.VenueRepository
}

func Newrepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) Repository {
	return Repository{
		OrganizerRepo: repository.NewOrganizerRepository(wrapDB, env),
		VenueRepo:     repository.NewVenueRepository(wrapDB, env),
	}
}
