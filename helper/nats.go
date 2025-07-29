package helper

import (
	"assist-tix/config"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

func CreateNatsConnection(env *config.EnvironmentVariable) (nc *nats.Conn, err error) {
	nc, err = nats.Connect(fmt.Sprintf("%v:%v", env.Nats.Addr, env.Nats.Port), nats.Token(env.Nats.Token))
	return
}

func CheckOrCreateNatsStream(js nats.JetStreamContext, streamName string, subjects []string) error {
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		// If the error is "nats: stream not found", it means the stream doesn't exist
		if err == nats.ErrStreamNotFound {

			// Define the stream configuration
			streamConfig := &nats.StreamConfig{
				Name:      streamName,
				Subjects:  subjects,
				Storage:   nats.FileStorage,
				Retention: nats.LimitsPolicy,
				MaxMsgs:   1000,
				MaxBytes:  1 * 1024 * 1024 * 1024, // 1GB
				MaxAge:    3600 * time.Second,     // 1 hour
				Replicas:  1,
			}

			// Add the stream
			_, err := js.AddStream(streamConfig)
			if err != nil {
				log.Fatal().Msgf("Error creating stream: %v", err)
			}
			log.Info().Msgf("Stream '%s' created successfully", streamName)
		} else {
			log.Error().Msgf("Error checking stream: %v", err)
			return err
		}
	} else {
		// Stream already exists
		log.Info().Msgf("Stream '%s' already exists with subjects: %v", stream.Config.Name, stream.Config.Subjects)
	}

	return nil
}
