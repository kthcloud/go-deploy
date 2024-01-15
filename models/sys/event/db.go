package event

import (
	"context"
	"time"
)

// Create creates a new event in the database.
func (client *Client) Create(id string, params *CreateParams) error {
	event := Event{
		ID:        id,
		CreatedAt: time.Now(),
		Type:      params.Type,
		Source:    params.Source,
		Metadata:  params.Metadata,
	}

	_, err := client.ResourceClient.Collection.InsertOne(context.TODO(), event)
	return err
}
