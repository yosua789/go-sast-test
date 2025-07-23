package repository

import (
	"assist-tix/config"
	"assist-tix/lib"
	"assist-tix/model"
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type TransactionReleaserRepository interface {
	PublishReleaseTransactionWithDelay(ctx context.Context, transactionId string) (err error)
}

type TransactionReleaserRepositoryImpl struct {
	Nats *nats.Conn
	Env  *config.EnvironmentVariable
}

func NewTransactionReleaserRepository(
	nats *nats.Conn,
	env *config.EnvironmentVariable,
) TransactionReleaserRepository {
	return &TransactionReleaserRepositoryImpl{
		Nats: nats,
		Env:  env,
	}
}

func (r *TransactionReleaserRepositoryImpl) PublishReleaseTransactionWithDelay(ctx context.Context, transactionId string) (err error) {
	js, err := r.Nats.JetStream()
	if err != nil {
		return
	}

	var data = model.ReleaseTransactionJob{
		TransactionID: transactionId,
		CreatedAt:     time.Now(),
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return
	}

	_, err = js.Publish(r.Env.Nats.Stream.ReleaseTransaction, bytes)
	if err != nil {
		return
	}

	return
}

func (r *TransactionReleaserRepositoryImpl) ParseData(data interface{}) (body []byte, err error) {
	if data == nil {
		return
	}

	val := reflect.ValueOf(data)
	if val.IsZero() {
		log.Error().Msg("failed to parse data")
		err = &lib.ErrorInternalServer
		return
	}

	defer func() {
		if e := recover(); e != nil {
			log.Error().Msg("failed to parse data")
			err = &lib.ErrorInternalServer
		}
	}()

	switch val.Kind() {
	case reflect.String:
		body = []byte(val.String())
	case reflect.Slice:
		body = val.Bytes()
	default:
		err = &lib.ErrorInternalServer
	}

	return
}

func (r *TransactionReleaserRepositoryImpl) PublishJob(js nats.JetStreamContext, subject string, data interface{}, metadata map[string]interface{}) error {
	body, err := r.ParseData(data)
	if err != nil {
		return err
	}

	msg := nats.NewMsg(subject)
	msg.Data = body

	js.PublishAsyncPending()
	_, err = js.PublishMsg(msg)
	if err != nil {
		return err
	}

	return nil
}
