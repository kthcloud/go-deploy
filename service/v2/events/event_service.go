package events

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/event_repo"
)

// Create creates a new event.
func (c *Client) Create(id string, params *model.EventCreateParams) error {
	return event_repo.New().Create(id, params)
}
