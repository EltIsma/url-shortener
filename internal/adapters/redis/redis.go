package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	redisLimiter "github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "urlShortener:"

type Redis struct {
	client redis.UniversalClient
	logger *slog.Logger
}

func New(hosts []string, password string, logger *slog.Logger) (*Redis, *redis_rate.Limiter, error) {
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    hosts,
		Password: password,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("error connection to Redis: %w", err)
	}

	rdb := redisLimiter.NewUniversalClient(&redisLimiter.UniversalOptions{
		Addrs:    hosts,
		Password: password,
	})

	// создаем rate limiter
	rateLimiter := redis_rate.NewLimiter(rdb)

	return &Redis{
		client: client,
		logger: logger,
	}, rateLimiter, nil
}

func (rdb *Redis) Close() {
	if err := rdb.client.Close(); err != nil {
		rdb.logger.Error("error closing connection:", slog.String("error", err.Error()))
	}
}

func (r *Redis) Set(ctx context.Context, key any, value any, ttl time.Duration) error {
	err := r.client.Set(ctx, keyPrefix+anyToString(key), value, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, key any) (value any, err error) {
	url, err := r.client.Get(ctx, keyPrefix+anyToString(key)).Result()
	if err != nil {
		return "", err
	}

	return url, nil
}

func anyToString(v any) string {
	return fmt.Sprintf("%v", v)
}
