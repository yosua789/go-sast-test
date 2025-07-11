package router

import (
	"assist-tix/config"
	"assist-tix/handler"
	"assist-tix/helper"
	"assist-tix/lib"
	"assist-tix/middleware"
	"errors"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Env                        *config.EnvironmentVariable
	OrganizerHandler           handler.OrganizerHandler
	VenueHandler               handler.VenueHandler
	EventHandler               handler.EventHandler
	EventTicketCategoryHandler handler.EventTicketCategoryHandler
	Middleware                 middleware.Middleware
}

func NewRouter(handler Handler) *gin.Engine {
	if handler.Env.App.Mode == lib.ModeProd {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	if handler.Env.Api.CorsEnable {
		router.Use(handler.Middleware.CORSMiddleware())
	}

	HelloWorld(router)

	var apiRouterGroupName = "/api"

	if handler.Env.Api.BasePath != "/" {
		apiRouterGroupName = path.Join(handler.Env.Api.BasePath, apiRouterGroupName)
	}

	// Api router
	apiRouterGroup := router.Group(apiRouterGroupName)

	// Manual Static file
	ServeStaticFile(apiRouterGroup, "public")

	if handler.Env.App.Debug {
		SwaggerRouter(apiRouterGroup)
	}

	RouterApiV1(handler.Env.App.Debug, handler, apiRouterGroup)

	return router

}

func ServeStaticFile(router *gin.RouterGroup, dir string) {
	r := router.Group("/public")

	r.GET("/*filepath", func(ctx *gin.Context) {
		filepath := ctx.Param("filepath")
		fullpath := dir + filepath
		_, err := os.Stat(fullpath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				ctx.AbortWithError(http.StatusNotFound, &lib.ErrorFileNotFound)
				return
			}

			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		fileExt := helper.GetFileExtension(filepath)
		var contentType = "image/jpeg"
		switch fileExt {
		case "png":
			contentType = "image/png"
		case "heic":
			contentType = "image/heic"
		}

		ctx.Header("Content-Disposition", "inline")
		ctx.Header("Content-Type", contentType)
		ctx.File(fullpath)
	})
}

func HelloWorld(router *gin.Engine) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, "Hello World")
	})
}

func SwaggerRouter(router *gin.RouterGroup) {
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
