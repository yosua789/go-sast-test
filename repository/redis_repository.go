package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository interface {
	SetState(ctx context.Context, state string, value string, expiration time.Duration) error
	GetState(ctx context.Context, state string) (string, error)
	DeleteState(ctx context.Context, state string) error
}
type RedisRepositoryImpl struct {
	client *redis.Client
}

// NewRedisRepository creates a new instance of RedisRepositoryImpl.
func NewRedisRepository(client *redis.Client) RedisRepository {
	return &RedisRepositoryImpl{
		client: client,
	}
}

// SetState sets a state value in Redis with an expiration time.
// epiration time in minute
func (r *RedisRepositoryImpl) SetState(ctx context.Context, state string, value string, expiration time.Duration) error {
	err := r.client.Set(ctx, state, value, expiration*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("could not set state in Redis: %w", err)
	}
	return nil
}

// GetState retrieves a state value from Redis.
func (r *RedisRepositoryImpl) GetState(ctx context.Context, state string) (string, error) {
	value, err := r.client.Get(ctx, state).Result()
	if err == redis.Nil {
		return "", err // State not found
	} else if err != nil {
		return "", redis.ErrClosed
	}
	return value, nil
}

// DeleteState removes a state from Redis.
func (r *RedisRepositoryImpl) DeleteState(ctx context.Context, state string) error {
	err := r.client.Del(ctx, state).Err()
	if err != nil {
		return fmt.Errorf("could not delete state from Redis: %w", err)
	}
	return nil
}
