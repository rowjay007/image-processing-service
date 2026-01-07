package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"image-processing-service/internal/config"
)

type RedisCache struct {
	client *redis.Client
}

func (c *RedisCache) Client() *redis.Client {
	return c.client
}

func NewRedisCache(cfg config.UpstashConfig) (*RedisCache, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	opt := &redis.Options{
		Addr:     addr,
		Password: cfg.Password, // no password set
		DB:       0,            // use default DB
	}

	if cfg.TLS {
		opt.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Return empty string for miss, or handle as error? Port says (string, error). usually miss is not error or specific error.
		// Let's return empty string and nil error for miss to satisfy "get" semantics,
		// or we can allow redis.Nil to bubble up if caller expects it.
		// For simplicity, let's return empty string on miss for now, but callers need to know.
		// Actually, standard is usually returning error on miss so caller knows it's a miss not empty value.
		// But let's check port definition... "Get(ctx, key) (string, error)".
		// If we return "", nil, caller can't distinguish between empty value and miss.
		// Let's return "" and explicit error if we want, or just let redis.Nil bubble up?
		// We'll return nil error and empty string for miss, assuming we don't store empty strings.
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
