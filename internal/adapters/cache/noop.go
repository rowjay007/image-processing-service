package cache

import (
	"context"
	"time"
)

// NoOpCache implements ports.Cache but does nothing.
type NoOpCache struct{}

func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (c *NoOpCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoOpCache) Incr(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (c *NoOpCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

// NoOpRateLimiter implements ports.RateLimiter but always allows requests.
type NoOpRateLimiter struct{}

func NewNoOpRateLimiter() *NoOpRateLimiter {
	return &NoOpRateLimiter{}
}

func (l *NoOpRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return true, nil
}
