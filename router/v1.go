package router

import (
	"github.com/gin-gonic/gin"
)

func RouterApiV1(debug bool, h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/v1")

	OrganizerRouter(h, r)
}

func OrganizerRouter(h Handler, rg *gin.RouterGroup) {
	r := rg.Group("/organizers")

	r.POST("", h.OrganizerHandler.Create)
	r.GET("", h.OrganizerHandler.GetAll)
	r.GET("/:organizerId", h.OrganizerHandler.GetByID)
}
