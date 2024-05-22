package key_value

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-deploy/pkg/db"
	"go-deploy/utils"
	"strings"
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

// Del deletes the given key.
func (client *Client) Del(key string) error {
	return client.RedisClient.Del(context.TODO(), key).Err()
}

// SetNX sets the value of the given key if it does not exist.
func (client *Client) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return client.RedisClient.SetNX(context.TODO(), key, value, expiration).Result()
}

// Incr increments the value of the given key.
func (client *Client) Incr(key string) error {
	return client.RedisClient.Incr(context.Background(), key).Err()
}

// Decr decrements the value of the given key.
func (client *Client) Decr(key string) error {
	return client.RedisClient.Decr(context.Background(), key).Err()
}

// SetUpExpirationListener sets up a listener for expired key events for every key beginning with the given prefix.
// It is non-blocking and will run in a separate goroutine.
func (client *Client) SetUpExpirationListener(prefix string, handler func(key string) error) error {
	go func() {
		pubsub := client.RedisClient.PSubscribe(context.TODO(), "__keyevent@0__:expired")
		defer func(channel *redis.PubSub) {
			err := channel.Close()
			if err != nil {
				return
			}
		}(pubsub)

		channel := pubsub.Channel()
		for msg := range channel {
			if msg.Payload == "" {
				continue
			}

			if !strings.HasPrefix(msg.Payload, prefix) {
				continue
			}

			err := handler(msg.Payload)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to handle expired key event for key %s. details: %w", msg.Payload, err))
				return
			}
		}
	}()
	return nil
}
