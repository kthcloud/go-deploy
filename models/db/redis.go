package db

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-deploy/pkg/config"
	"log"
)

func (dbCtx *Context) setupRedis() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup redis. details: %w", err)
	}

	log.Println("setting up redis")

	dbCtx.RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.URL,
		Password: config.Config.Redis.Password,
		DB:       0, // use default DB
	})

	_, err := dbCtx.RedisClient.Ping(context.TODO()).Result()
	if err != nil {
		return makeError(err)
	}

	log.Println("connected to redis")

	return nil
}

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
