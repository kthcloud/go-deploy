package events

import "go-deploy/models/sys/event"

// Create creates a new event.
func (c *Client) Create(id string, params *event.CreateParams) error {
	return event.New().Create(id, params)
}
