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

	r.POST("", h.Middleware.CORSMiddleware(), h.OrganizerHandler.Create)
	r.GET("", h.Middleware.CORSMiddleware(), h.OrganizerHandler.GetAll)
	r.GET("/:organizerId", h.Middleware.CORSMiddleware(), h.OrganizerHandler.GetByID)
	r.PUT("/:organizerId", h.Middleware.CORSMiddleware(), h.OrganizerHandler.Update)
	r.DELETE("/:organizerId", h.Middleware.CORSMiddleware(), h.OrganizerHandler.Delete)
}

func VenueRouter(h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/venues")

	r.POST("", h.Middleware.CORSMiddleware(), h.VenueHandler.Create)
	r.GET("", h.Middleware.CORSMiddleware(), h.VenueHandler.GetAll)
	r.GET("/:venueId", h.Middleware.CORSMiddleware(), h.VenueHandler.GetById)
	r.PUT("/:venueId", h.Middleware.CORSMiddleware(), h.VenueHandler.Update)
	r.DELETE("/:venueId", h.Middleware.CORSMiddleware(), h.VenueHandler.Delete)
}

func EventRouter(h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/events")

	r.GET("", h.Middleware.CORSMiddleware(), h.EventHandler.GetAllPaginated)
	r.GET("/:eventId", h.Middleware.CORSMiddleware(), h.EventHandler.GetById)
	r.DELETE("/:eventId", h.Middleware.CORSMiddleware(), h.EventHandler.Delete)

	EventTicketCategories(h, r)
}

func EventTicketCategories(h Handler, rg *gin.RouterGroup) {
	// /events/{eventId}/ticket-categories

	rg.POST("/:eventId/ticket-categories", h.Middleware.CORSMiddleware(), h.EventTicketCategoryHandler.Create)
	rg.GET("/:eventId/ticket-categories", h.Middleware.CORSMiddleware(), h.EventTicketCategoryHandler.GetByEventId)
	rg.GET("/:eventId/ticket-categories/:ticketCategoryId", h.Middleware.CORSMiddleware(), h.EventTicketCategoryHandler.GetById)
}
