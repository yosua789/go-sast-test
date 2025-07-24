package api

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/helper"
	"assist-tix/middleware"
	"assist-tix/router"
	"assist-tix/storage"
	custValidator "assist-tix/validator"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type Setup struct {
	Router     *gin.Engine
	Service    Service
	Repository Repository
	WrapDB     *database.WrapDB
}

func Init(env *config.EnvironmentVariable) (*Setup, error) {
	wrapDB := database.InitDB(env)

	// Init GCS
	gcsClient, err := storage.NewGCSClient(env)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect gcs")
		return nil, err
	}
	env.Paylabs.PrivateKey = helper.GetKeyFileString(env.Paylabs.PrivateKey)
	env.Paylabs.PublicKey = helper.GetKeyFileString(env.Paylabs.PublicKey)

	validate := validator.New()
	custValidator.InitCustomValidator(validate)

	// Init asynq
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: env.Redis.Host, Username: env.Redis.Username, Password: env.Redis.Password})
	err = asynqClient.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("asynq didn't respond")
	}
	job := NewJob(env, asynqClient)

	repository := Newrepository(wrapDB, env, gcsClient)
	service := Newservice(env, repository, wrapDB, job)
	handler := Newhandler(env, service, validate)

	middleware := middleware.NewMiddleware(env)

	r := router.Handler{
		Env:                        env,
		OrganizerHandler:           handler.OrganizerHandler,
		VenueHandler:               handler.VenueHandler,
		SectorHandler:              handler.SectorHandler,
		EventHandler:               handler.EventHandler,
		EventTicketCategoryHandler: handler.EventTicketCategoryHandler,
		EventTransaction:           handler.EventTransactionHandler,
		Middleware:                 middleware,
	}

	routes := router.NewRouter(r)

	return &Setup{
		Router:     routes,
		Repository: repository,
		Service:    service,
		WrapDB:     wrapDB,
	}, nil
}
