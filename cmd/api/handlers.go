package api

import (
	"assist-tix/config"
	"assist-tix/handler"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	OrganizerHandler           handler.OrganizerHandler
	VenueHandler               handler.VenueHandler
	SectorHandler              handler.SectorHandler
	EventHandler               handler.EventHandler
	EventTicketCategoryHandler handler.EventTicketCategoryHandler
	EventTransactionHandler    handler.EventTransactionHandler
}

func Newhandler(
	env *config.EnvironmentVariable,
	s Service,
	validator *validator.Validate,
) Handler {
	return Handler{
		OrganizerHandler:           handler.NewOrganizerHandler(env, s.OrganizerService, validator),
		VenueHandler:               handler.NewVenueHandler(env, s.VenueService, validator),
		SectorHandler:              handler.NewSectorHandler(env, s.VenueService, validator),
		EventHandler:               handler.NewEventHandler(env, s.EventService, validator),
		EventTicketCategoryHandler: handler.NewEventTicketCategoryHandler(env, s.EventTicketCategoryService, validator),
		EventTransactionHandler:    handler.NewEventTransactionHandler(env, s.EventTransactionService, validator),
	}
}
