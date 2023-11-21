package key_value

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"go-deploy/models/db"
	"time"
)

type Client struct {
	RedisClient *redis.Client
}

func New() *Client {
	return &Client{
		RedisClient: db.DB.RedisClient,
	}
}

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

func (client *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return client.RedisClient.Set(context.TODO(), key, value, expiration).Err()
}

func (client *Client) Incr(key string) error {
	return client.RedisClient.Incr(context.Background(), key).Err()
}

func (client *Client) Decr(key string) error {
	return client.RedisClient.Decr(context.Background(), key).Err()
}
