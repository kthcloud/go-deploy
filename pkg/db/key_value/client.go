package key_value

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Client is used to manage key-value pairs in Redis.
type Client struct {
	RedisClient *redis.Client
	Logger      *zap.Logger
}

// New returns a new key-value client backed by the global Redis client.
func New() *Client {
	return NewWithClient(db.DB.RedisClient, nil)
}

// NewWithClient constructs a Client with an explicit Redis client and optional logger.
func NewWithClient(redisClient *redis.Client, logger *zap.Logger) *Client {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Client{
		RedisClient: redisClient,
		Logger:      logger,
	}
}

func (client *Client) log() *zap.Logger {
	if client.Logger == nil {
		return zap.NewNop()
	}
	return client.Logger
}

// Get returns the value of the given key, or "" if the key does not exist.
func (client *Client) Get(ctx context.Context, key string) (string, error) {
	res, err := client.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}

		return "", err
	}

	return res, nil
}

// List returns all keys that match the given pattern.
func (client *Client) List(ctx context.Context, pattern string) ([]string, error) {
	keys, err := client.RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// Set sets the value of the given key.
func (client *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return client.RedisClient.Set(ctx, key, value, expiration).Err()
}

// Del deletes the given key.
func (client *Client) Del(ctx context.Context, key string) error {
	return client.RedisClient.Del(ctx, key).Err()
}

// SetNX sets the value of the given key if it does not exist.
func (client *Client) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return client.RedisClient.SetNX(ctx, key, value, expiration).Result()
}

// SetXX sets the value of the given key if it exists.
func (client *Client) SetXX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return client.RedisClient.SetXX(ctx, key, value, expiration).Result()
}

// IsSet checks if the given key is set.
func (client *Client) IsSet(ctx context.Context, key string) (bool, error) {
	res, err := client.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return res > 0, nil
}

// Incr increments the value of the given key.
func (client *Client) Incr(ctx context.Context, key string) error {
	return client.RedisClient.Incr(ctx, key).Err()
}

// Decr decrements the value of the given key.
func (client *Client) Decr(ctx context.Context, key string) error {
	return client.RedisClient.Decr(ctx, key).Err()
}

// SetUpExpirationListener sets up a listener for expired key events matching pattern.
// It is non-blocking and exits when ctx is canceled.
func (client *Client) SetUpExpirationListener(ctx context.Context, pattern string, handler func(key string) error) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	go func() {
		pubSub := client.RedisClient.PSubscribe(ctx, "__keyevent@0__:expired")
		defer func() {
			_ = pubSub.Close()
		}()

		channel := pubSub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-channel:
				if !ok || msg == nil || msg.Payload == "" {
					if !ok {
						return
					}
					continue
				}

				if !re.MatchString(msg.Payload) {
					continue
				}

				if err := handler(msg.Payload); err != nil {
					client.log().Warn("failed to handle expired key event", zap.String("key", msg.Payload), zap.Error(err))
				}
			}
		}
	}()

	return nil
}
