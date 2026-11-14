package db

import (
	"context"
	"fmt"

	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/redis/go-redis/v9"
)

// setupRedis initializes the Redis connection.
// It should be called once at the start of the application.
func (dbCtx *Context) setupRedis(ctx context.Context) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to set up redis: %w", err)
	}

	log.Println("Setting up Redis")

	dbCtx.RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.URL,
		Password: config.Config.Redis.Password,
		DB:       0, // use default DB
	})

	if _, err := dbCtx.RedisClient.Ping(ctx).Result(); err != nil {
		return makeError(err)
	}

	if err := dbCtx.RedisClient.ConfigSet(ctx, "notify-keyspace-events", "Ex").Err(); err != nil {
		return makeError(err)
	}

	return nil
}

// shutdownRedis closes the Redis connection.
// It should be called once at the end of the application.
func (dbCtx *Context) shutdownRedis() error {
	if dbCtx.RedisClient == nil {
		return nil
	}

	if err := dbCtx.RedisClient.Close(); err != nil {
		return fmt.Errorf("failed to shutdown redis: %w", err)
	}

	dbCtx.RedisClient = nil

	return nil
}
