package db

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
)

// setupRedis initializes the Redis connection.
// It should be called once at the start of the application.
func (dbCtx *Context) setupRedis() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to set up redis. details: %w", err)
	}

	log.Println("Setting up Redis")

	dbCtx.RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.URL,
		Password: config.Config.Redis.Password,
		DB:       0, // use default DB
	})

	_, err := dbCtx.RedisClient.Ping(context.TODO()).Result()
	if err != nil {
		return makeError(err)
	}

	err = dbCtx.RedisClient.ConfigSet(context.TODO(), "notify-keyspace-events", "Ex").Err()
	if err != nil {
		return makeError(err)
	}

	return nil
}

// shutdownRedis closes the Redis connection.
// It should be called once at the end of the application.
func (dbCtx *Context) shutdownRedis() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to shutdown redis. details: %w", err)
	}

	err := dbCtx.RedisClient.Close()
	if err != nil {
		return makeError(err)
	}

	dbCtx.RedisClient = nil

	return nil
}
