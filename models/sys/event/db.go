package event

import (
	"context"
	"time"
)

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
