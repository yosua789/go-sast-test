package mailer

import (
	"assist-tix/config"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

func Init(env *config.EnvironmentVariable) *Nats {
	natsString := fmt.Sprintf("%v:%v", env.Nats.Addr, env.Nats.Port)
	nc, err := nats.Connect(natsString, nats.Token(env.Nats.Token))
	if err != nil {
		log.Fatal().Err(err).Msg("[x] failed to connect to nats")
		panic(err)
	}
	log.Info().Msgf("[+] Successfully connected to NATS at %s", natsString)

	js, err := nc.JetStream()
	if err != nil {
		log.Fatal().Err(err).Msg("[x] failed to create jetstream")
		panic(err)
	}
	log.Info().Msg("[+] Successfully created JetStream")

	m := NewNats(js)
	return m
}
