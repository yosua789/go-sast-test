package router

import (
	"github.com/gin-gonic/gin"
)

func RouterApiV1(debug bool, h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/v1")

	OrganizerRouter(h, r)
	VenueRouter(h, r)
	EventRouter(h, r)
	ExternalRouter(h, r)
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
	r.GET("/:eventId/active-settings", h.EventHandler.GetActiveSettings)
	r.DELETE("/:eventId", h.EventHandler.Delete)
	r.GET("/:eventId/verify/garuda-id/:garudaId", h.EventHandler.VerifyGarudaID)

	// Validate book email
	r.GET("/:eventId/email-books/:email", h.EventTransaction.IsEmailAlreadyBook)
	r.GET("/:eventId/payment-methods", h.EventTransaction.GetAvailablePaymentMethods)

	r.GET("/transactions/:transactionId", h.Middleware.TokenAuthMiddleware(), h.EventTransaction.GetTransactionDetails)

	r.GET("/transactions/:transactionId", h.Middleware.TokenAuthMiddleware(), h.EventTransaction.GetTransactionDetails)

	EventTicketCategories(h, r)
}

func EventTicketCategories(h Handler, rg *gin.RouterGroup) {
	// /events/{eventId}/ticket-categories

	rg.POST("/:eventId/ticket-categories", h.EventTicketCategoryHandler.Create)
	rg.GET("/:eventId/ticket-categories", h.EventTicketCategoryHandler.GetByEventId)
	rg.GET("/:eventId/ticket-categories/:ticketCategoryId", h.EventTicketCategoryHandler.GetById)

	rg.GET("/:eventId/ticket-categories/:ticketCategoryId/seatmap", h.EventTicketCategoryHandler.GetSeatmap)

	rg.POST("/:eventId/ticket-categories/:ticketCategoryId/order", h.EventTransaction.CreateTransaction)
	rg.POST("/:eventId/ticket-categories/:ticketCategoryId/order/paylabs-vasnap", h.EventTransaction.PaylabsVASnap)
}

func ExternalRouter(h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/external")

	r.POST("/paylabs/va-snap/callback", h.Middleware.PayloadPasser(), h.EventTransaction.CallbackVASnap)
}
