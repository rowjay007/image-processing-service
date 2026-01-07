package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

func (rl *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {

	pipe := rl.client.Pipeline()
	incr := pipe.Incr(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := incr.Val()
	if count == 1 {
		rl.client.Expire(ctx, key, window)
	}

	if count > int64(limit) {
		return false, nil
	}

	return true, nil
}
