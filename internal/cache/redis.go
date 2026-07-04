// Package cache is a Redis connector that stores the JWT tokens minted for
// service accounts, keyed by account id.
package cache

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound is returned by GetServiceToken when no token is cached for the
// account (missing or expired).
var ErrNotFound = errors.New("cache: token not found")

// Config holds the settings needed to reach a Redis instance.
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Addr renders the host:port Redis address.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

// Client is a thin wrapper over the Redis client scoped to this app's needs.
type Client struct {
	rdb *redis.Client
}

// New opens a Redis connection and verifies it with a ping.
func New(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close releases the underlying connection pool.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// serviceTokenKey is the Redis key holding a service account's current token.
func serviceTokenKey(accountID int32) string {
	return fmt.Sprintf("service:token:%d", accountID)
}

// SetServiceToken stores (or replaces) the token for a service account, expiring
// it after ttl so the cached copy dies with the JWT.
func (c *Client) SetServiceToken(ctx context.Context, accountID int32, token string, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, serviceTokenKey(accountID), token, ttl).Err(); err != nil {
		return fmt.Errorf("set service token: %w", err)
	}
	return nil
}

// GetServiceToken returns the cached token for a service account, or ErrNotFound
// if none is stored.
func (c *Client) GetServiceToken(ctx context.Context, accountID int32) (string, error) {
	token, err := c.rdb.Get(ctx, serviceTokenKey(accountID)).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return "", ErrNotFound
	case err != nil:
		return "", fmt.Errorf("get service token: %w", err)
	}
	return token, nil
}

// DeleteServiceToken removes the cached token for a service account.
func (c *Client) DeleteServiceToken(ctx context.Context, accountID int32) error {
	if err := c.rdb.Del(ctx, serviceTokenKey(accountID)).Err(); err != nil {
		return fmt.Errorf("delete service token: %w", err)
	}
	return nil
}
