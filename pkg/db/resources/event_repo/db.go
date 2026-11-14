package event_repo

import (
	"github.com/kthcloud/go-deploy/models/model"
	"time"
)

// Create creates a new event in the database.
func (client *Client) Create(id string, params *model.EventCreateParams) error {
	event := model.Event{
		ID:        id,
		CreatedAt: time.Now(),
		Type:      params.Type,
		Source:    params.Source,
		Metadata:  params.Metadata,
	}

	_, err := client.Collection.InsertOne(client.Context(), event)
	return err
}
