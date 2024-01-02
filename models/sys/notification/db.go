package notification

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (client *Client) Create(id string, userID string, params *CreateParams) (*Notification, error) {
	notification := &Notification{
		ID:          id,
		UserID:      userID,
		Type:        params.Type,
		Content:     params.Content,
		CreatedAt:   time.Now(),
		ReadAt:      time.Time{},
		CompletedAt: time.Time{},
		DeletedAt:   time.Time{},
	}

	_, err := client.Collection.InsertOne(context.TODO(), notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create job. details: %w", err)
	}

	return notification, nil
}

func (client *Client) MarkCompletedByID(id string) error {
	return client.SetWithBsonByID(id, bson.D{{"completedAt", time.Now()}})
}

func (client *Client) MarkReadByID(id string) error {
	return client.SetWithBsonByID(id, bson.D{{"readAt", time.Now()}})
}

func (client *Client) MarkReadAndCompleted() error {
	return client.SetWithBson(bson.D{{"readAt", time.Now()}, {"completedAt", time.Now()}})
}
