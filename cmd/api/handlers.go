package api

import (
	"assist-tix/config"
	"assist-tix/handler"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	OrganizerHandler handler.OrganizerHandler
}

func Newhandler(
	env *config.EnvironmentVariable,
	s Service,
	validator *validator.Validate,
) Handler {
	return Handler{
		OrganizerHandler: handler.NewOrganizerHandler(env, s.OrganizerService, validator),
	}
}
