package notification_repo

import (
	"context"
	"fmt"
	"go-deploy/models/model"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (client *Client) Create(id string, userID string, params *model.NotificationCreateParams) (*model.Notification, error) {
	notification := &model.Notification{
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

func (client *Client) MarkToastedByID(id string) error {
	return client.SetWithBsonByID(id, bson.D{{"toastedAt", time.Now()}})
}

func (client *Client) MarkReadAndCompleted() error {
	return client.SetWithBSON(bson.D{{"readAt", time.Now()}, {"completedAt", time.Now()}})
}
