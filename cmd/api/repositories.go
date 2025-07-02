package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/repository"
)

type Repository struct {
	OrganizerRepo repository.OrganizerRepository
}

func Newrepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) Repository {
	return Repository{
		OrganizerRepo: repository.NewOrganizerRepository(wrapDB, env),
	}
}
