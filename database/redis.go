package database

import (
	"context"
	"fmt"
	"go-realtime-workspace/config"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the redis.Client connection.
type RedisClient struct {
	*redis.Client
	cfg config.RedisConfig
}

// NewRedisClient creates a new Redis client connection.
func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:   cfg.Password,
		DB:         cfg.DB,
		MaxRetries: cfg.MaxRetries,
		PoolSize:   cfg.PoolSize,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	return &RedisClient{
		Client: client,
		cfg:    cfg,
	}, nil
}

// GetConfig returns the Redis configuration.
func (r *RedisClient) GetConfig() config.RedisConfig {
	return r.cfg
}

// HealthCheck performs a Redis health check.
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	return r.Ping(ctx).Err()
}
