package notification

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (client *Client) Create(id string, userID string, params *CreateParams) (*Notification, error) {
	notification := &Notification{
		ID:        id,
		UserID:    userID,
		Type:      params.Type,
		Content:   params.Content,
		CreatedAt: time.Now(),
		ReadAt:    nil,
		DeletedAt: nil,
	}

	_, err := client.Collection.InsertOne(context.TODO(), notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create job. details: %w", err)
	}

	return notification, nil
}

func (client *Client) UpdateWithParamsByID(id string, params *UpdateParams) error {
	update := bson.D{}

	models.AddIfNotNil(&update, "readAt", params.ReadAt)

	return client.UpdateWithBsonByID(id, update)
}
