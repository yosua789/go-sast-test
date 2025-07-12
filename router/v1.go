package router

import (
	"github.com/gin-gonic/gin"
)

func RouterApiV1(debug bool, h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/v1")

	OrganizerRouter(h, r)
	VenueRouter(h, r)
	EventRouter(h, r)
}

func OrganizerRouter(h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/organizers")

	r.POST("", h.OrganizerHandler.Create)
	r.GET("", h.OrganizerHandler.GetAll)
	r.GET("/:organizerId", h.OrganizerHandler.GetByID)
	r.PUT("/:organizerId", h.OrganizerHandler.Update)
	r.DELETE("/:organizerId", h.OrganizerHandler.Delete)
}

func VenueRouter(h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/venues")

	r.POST("", h.VenueHandler.Create)
	r.GET("", h.VenueHandler.GetAll)
	r.GET("/:venueId", h.VenueHandler.GetById)
	r.PUT("/:venueId", h.VenueHandler.Update)
	r.DELETE("/:venueId", h.VenueHandler.Delete)

	r.GET("/:venueId/sectors", h.SectorHandler.GetByVenueId)
}

func EventRouter(h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/events")

	r.GET("", h.EventHandler.GetAllPaginated)
	r.GET("/:eventId", h.EventHandler.GetById)
	r.DELETE("/:eventId", h.EventHandler.Delete)

	EventTicketCategories(h, r)
}

func EventTicketCategories(h Handler, rg *gin.RouterGroup) {
	// /events/{eventId}/ticket-categories

	rg.POST("/:eventId/ticket-categories", h.EventTicketCategoryHandler.Create)
	rg.GET("/:eventId/ticket-categories", h.EventTicketCategoryHandler.GetByEventId)
	rg.GET("/:eventId/ticket-categories/:ticketCategoryId", h.EventTicketCategoryHandler.GetById)

	rg.GET("/:eventId/ticket-categories/:ticketCategoryId/seatmap", h.EventTicketCategoryHandler.GetSeatmap)
}
