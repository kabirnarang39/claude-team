package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps a redis.Client with domain key builders and fail-open helpers.
type Client struct {
	rdb *redis.Client
}

// New parses redisURL and returns a connected Client.
func New(redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("cache: parse redis URL: %w", err)
	}
	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: ping redis: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

// Client returns the underlying redis.Client.
func (c *Client) Client() *redis.Client {
	return c.rdb
}

// Close closes the underlying connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Key builders —————————————————————————————————————————

// KeyPlanLimits returns the cache key for a workspace's plan limits.
func KeyPlanLimits(workspaceID string) string {
	return "plan_limits:" + workspaceID
}

// KeyUsage returns the cache key for API usage in a period (e.g. "2026-06").
func KeyUsage(workspaceID, period string) string {
	return "billing:" + workspaceID + ":" + period
}

// KeyRateLimit returns the per-user per-minute rate limit counter key.
func KeyRateLimit(userID, minute string) string {
	return "ratelimit:" + userID + ":" + minute
}

// KeyIdempotency returns the idempotency result cache key.
func KeyIdempotency(userID, key string) string {
	return "idem:" + userID + ":" + key
}

// Fail-open helpers ————————————————————————————————————

// GetFailOpen returns the string value for key, or "" on any error.
// Use for non-critical reads where a cache miss is acceptable.
func (c *Client) GetFailOpen(ctx context.Context, key string) string {
	v, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			log.Printf("cache: get fail-open key=%s err=%v", key, err)
		}
		return ""
	}
	return v
}

// IncrFailOpen increments key and returns the new value, or 0 on error.
func (c *Client) IncrFailOpen(ctx context.Context, key string) int64 {
	n, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("cache: incr fail-open key=%s err=%v", key, err)
		return 0
	}
	return n
}

// Regular operations ———————————————————————————————————

// Get returns the string value for key, or an error (including redis.Nil for missing).
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	v, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return v, nil
}

// Set stores value with the given TTL. Pass 0 for no expiry.
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// SetNX sets key to value with ttl only if it does not already exist.
// Returns true if the key was set.
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, key, value, ttl).Result()
}

// Del deletes one or more keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Incr atomically increments the integer value of key by 1.
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// Expire sets a TTL on an existing key.
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}
