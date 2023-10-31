package key_value

import (
	"context"
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
	return client.RedisClient.Get(context.TODO(), key).Result()
}

func (client *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return client.RedisClient.Set(context.TODO(), key, value, expiration).Err()
}
