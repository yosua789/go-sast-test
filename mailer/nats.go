package mailer

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/nats-io/nats.go"
)

var (
	errZeroValue   = errors.New("[NatsParseData]: can't accept zero value")
	errInvalidType = errors.New("[NatsParseData]: can't accept other than string type or slice of byte type")
)

type Nats struct {
	JS nats.JetStreamContext
}

func NewNats(JS nats.JetStreamContext) *Nats {
	return &Nats{JS: JS}
}

func (n *Nats) Publish(subject string, data interface{}, metadata map[string]interface{}) error {
	body, err := n.parseData(data)

	header := n.parseMetadata(metadata)
	if err != nil {
		return err
	}
	msg := nats.NewMsg(subject)
	msg.Data = body

	for k, v := range header {
		msg.Header.Set(k, v)
	}

	n.JS.PublishAsyncPending()
	_, err = n.JS.PublishMsg(msg)
	if err != nil {
		return err
	}

	return nil
}

func (n *Nats) parseData(data interface{}) (body []byte, err error) {
	if data == nil {
		return
	}

	val := reflect.ValueOf(data)
	if val.IsZero() {
		err = errZeroValue
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = errInvalidType
		}
	}()

	switch val.Kind() {
	case reflect.String:
		body = []byte(val.String())
	case reflect.Slice:
		body = val.Bytes()
	default:
		err = errInvalidType
	}
	return
}

func (n *Nats) parseMetadata(m map[string]interface{}) map[string]string {
	parsedMetadata := make(map[string]string)
	if m == nil {
		return parsedMetadata
	}

	for k, v := range m {
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.String:
			parsedMetadata[k] = val.String()
		case reflect.Int:
			parsedMetadata[k] = strconv.Itoa(int(val.Int()))
		case reflect.Bool:
			parsedMetadata[k] = strconv.FormatBool(val.Bool())
		default:
			continue
		}
	}

	return parsedMetadata
}
