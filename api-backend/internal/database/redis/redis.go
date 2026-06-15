package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Database struct {
	client *goredis.Client
}

func New(client *goredis.Client) *Database {
	return &Database{client: client}
}

func (d *Database) Get(ctx context.Context, key string) (string, error) {
	value, err := d.client.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return "", err
	}
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", goredis.Nil
	}

	return value, nil
}

func (d *Database) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return d.client.Set(ctx, key, value, ttl).Err()
}
