package api

import (
	"assist-tix/config"
	"assist-tix/handler"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	OrganizerHandler           handler.OrganizerHandler
	VenueHandler               handler.VenueHandler
	EventHandler               handler.EventHandler
	EventTicketCategoryHandler handler.EventTicketCategoryHandler
}

func Newhandler(
	env *config.EnvironmentVariable,
	s Service,
	validator *validator.Validate,
) Handler {
	return Handler{
		OrganizerHandler:           handler.NewOrganizerHandler(env, s.OrganizerService, validator),
		VenueHandler:               handler.NewVenueHandler(env, s.VenueService, validator),
		EventHandler:               handler.NewEventHandler(env, s.EventService, validator),
		EventTicketCategoryHandler: handler.NewEventTicketCategoryHandler(env, s.EventTicketCategoryService, validator),
	}
}
