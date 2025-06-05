package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-kaas/api/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Client wraps the Redis client
type Client struct {
	client *redis.Client
	logger *zap.Logger
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig, logger *zap.Logger) (*Client, error) {
	opt := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.Int("db", cfg.DB))

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// SetWithTTL sets a key with TTL
func (c *Client) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Get gets a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	return val, err
}

// Delete deletes a key
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// SetJSON sets a JSON value with TTL
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// GetDel gets and deletes a key atomically
func (c *Client) GetDel(ctx context.Context, key string) (string, error) {
	val, err := c.client.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	return val, err
}