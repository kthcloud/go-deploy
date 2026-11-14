package message_queue

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	// ErrQueueNotFound is returned when a queue is not found.
	ErrQueueNotFound = errors.New("queue not found")
)

// Client is used to publish and consume messages in Redis Pub/Sub.
type Client struct {
	RedisClient *redis.Client
	Logger      *zap.Logger
}

// New returns a new message queue client backed by the global Redis client.
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

// Publish sends a message to the given queue.
func (client *Client) Publish(ctx context.Context, queueName string, jsonData any) error {
	data, err := json.Marshal(jsonData)
	if err != nil {
		client.log().Warn("failed to marshal message", zap.String("queue", queueName), zap.Error(err))
		return err
	}

	return client.RedisClient.Publish(ctx, queueName, data).Err()
}

// Consume starts consuming messages from the given queue.
// It is non-blocking and exits when ctx is canceled or the subscription closes.
func (client *Client) Consume(ctx context.Context, queueName string, handler func(data []byte) error) error {
	go func() {
		pubSub := client.RedisClient.Subscribe(ctx, queueName)
		defer func() {
			if err := pubSub.Close(); err != nil {
				client.log().Warn("failed to close pubsub channel", zap.String("queue", queueName), zap.Error(err))
			}
		}()

		channel := pubSub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-channel:
				if !ok || msg == nil {
					if !ok {
						return
					}
					continue
				}

				if err := handler([]byte(msg.Payload)); err != nil {
					client.log().Warn("failed to handle message", zap.String("queue", queueName), zap.Error(err))
				}
			}
		}
	}()

	return nil
}

// GetListeners returns the number of listeners for the given queue.
func (client *Client) GetListeners(ctx context.Context, queueName string) (int, error) {
	res, err := client.RedisClient.PubSubNumSub(ctx, queueName).Result()
	if err != nil {
		client.log().Warn("failed to get number of listeners", zap.String("queue", queueName), zap.Error(err))
		return 0, err
	}

	if len(res) == 0 {
		return 0, nil
	}

	return int(res[queueName]), nil
}
