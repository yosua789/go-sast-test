package domain

import (
	"context"
)

type EventPublisher interface {
	Publish(ctx context.Context, subject string, data []byte) (err error)
}
