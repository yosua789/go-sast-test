package main

import (
	"assist-tix/cmd/api"
	"assist-tix/config"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

// @title Assist Tix Documentation
// @version 1.0
// @description Assist Tix API Documentation
// @host localhost:3000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	env, err := config.LoadEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
		panic(err)
	}
	config.InitSwagger(env)
	config.InitLogger(env)

	setup, err := api.Init(env)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to initialize service")
		panic(err)
	}

	defer setup.WrapDB.Postgres.Conn.Close()

	go func() {
		err = setup.Router.Run(env.App.Host)
		if err != nil {
			log.Info().Msg(fmt.Sprintf("Listening on %s", env.App.Host))
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	<-done
	close(done)
	log.Info().Msg("Assist Tix API exited properly")
}
