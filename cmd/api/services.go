package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/service"
)

type Service struct {
	OrganizerService service.OrganizerService
}

func Newservice(
	env *config.EnvironmentVariable,
	r Repository,
	db *database.WrapDB,
) Service {
	organizerService := service.NewOrganizerService(db, env, r.OrganizerRepo)
	return Service{
		OrganizerService: organizerService,
	}
}
