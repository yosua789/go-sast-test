package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/middleware"
	"assist-tix/router"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Setup struct {
	Router     *gin.Engine
	Service    Service
	Repository Repository
	WrapDB     *database.WrapDB
}

func Init(env *config.EnvironmentVariable) (*Setup, error) {

	wrapDB := database.InitDB(env)

	repository := Newrepository(wrapDB, env)

	service := Newservice(env, repository, wrapDB)

	validate := validator.New()

	handler := Newhandler(env, service, validate)

	middleware := middleware.NewMiddleware(env)

	r := router.Handler{
		Env:              env,
		OrganizerHandler: handler.OrganizerHandler,
		VenueHandler:     handler.VenueHandler,
		EventHandler:     handler.EventHandler,
		Middleware:       middleware,
	}

	routes := router.NewRouter(r)

	return &Setup{
		Router:     routes,
		Repository: repository,
		Service:    service,
		WrapDB:     wrapDB,
	}, nil
}
