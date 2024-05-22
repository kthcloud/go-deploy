package message_queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-deploy/pkg/db"
	"go-deploy/pkg/log"
)

var (
	// QueueNotFoundErr is returned when a queue is not found.
	QueueNotFoundErr = fmt.Errorf("queue not found")
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

// Publish sends a message to the given queue.
func (client *Client) Publish(queueName string, jsonData interface{}) error {
	data, err := json.Marshal(jsonData)
	if err != nil {
		log.Println("Failed to marshal message. Details: " + err.Error())
		return err
	}

	return client.RedisClient.Publish(context.TODO(), queueName, data).Err()
}

// Consume starts consuming messages from the given queue.
// It is non-blocking and will run in a separate goroutine.
func (client *Client) Consume(ctx context.Context, queueName string, handler func(data []byte) error) error {
	go func() {
		pubSub := client.RedisClient.Subscribe(context.TODO(), queueName)
		defer func(channel *redis.PubSub) {
			err := channel.Close()
			if err != nil {
				log.Println("Failed to close channel. Details: " + err.Error())
				return
			}
		}(pubSub)

		channel := pubSub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-channel:
				err := handler([]byte(msg.Payload))
				if err != nil {
					log.Println("Failed to handle message. Details: " + err.Error())
					continue
				}
			}
		}
	}()

	return nil
}

// GetListeners returns the number of listeners for the given queue.
func (client *Client) GetListeners(queueName string) (int, error) {
	res, err := client.RedisClient.PubSubNumSub(context.TODO(), queueName).Result()
	if err != nil {
		log.Println("Failed to get number of listeners. Details: " + err.Error())
		return 0, err
	}

	if len(res) == 0 {
		return 0, nil
	}

	return int(res[queueName]), nil
}
