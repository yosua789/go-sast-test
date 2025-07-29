package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Publisher struct {
	NatsConn  *nats.Conn
	Jetstream jetstream.JetStream
}

func NewPublisher(
	conn *nats.Conn,
	jetStream jetstream.JetStream,
) *Publisher {
	return &Publisher{
		NatsConn:  conn,
		Jetstream: jetStream,
	}
}

func (p *Publisher) Publish(ctx context.Context, subject string, data []byte) (err error) {
	_, err = p.Jetstream.Publish(ctx, subject, data)
	return
}
