package events

import (
	"context"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/event_repo"
)

// Create creates a new event.
func (c *Client) Create(ctx context.Context, id string, params *model.EventCreateParams) error {
	erc := event_repo.New()
	erc.Ctx = ctx
	return erc.Create(id, params)
}
