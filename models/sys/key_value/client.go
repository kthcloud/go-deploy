package key_value

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"go-deploy/models/db"
	"time"
)

// Client is used to manage key-value pairs in Redis.
type Client struct {
	RedisClient *redis.Client
}

// New returns a new key-value client.
func New() *Client {
	return &Client{
		RedisClient: db.DB.RedisClient,
	}
}

// Get returns the value of the given key.
func (client *Client) Get(key string) (string, error) {
	res, err := client.RedisClient.Get(context.TODO(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}

		return "", err
	}

	return res, err
}

// Set sets the value of the given key.
func (client *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return client.RedisClient.Set(context.TODO(), key, value, expiration).Err()
}

// Incr increments the value of the given key.
func (client *Client) Incr(key string) error {
	return client.RedisClient.Incr(context.Background(), key).Err()
}

// Decr decrements the value of the given key.
func (client *Client) Decr(key string) error {
	return client.RedisClient.Decr(context.Background(), key).Err()
}
