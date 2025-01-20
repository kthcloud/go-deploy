package notifications

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/notification_repo"
	"github.com/kthcloud/go-deploy/service/clients"
	"github.com/kthcloud/go-deploy/service/core"
	"sort"
)

// Client is the client for the notification service.
type Client struct {
	// V2 is a reference to the parent client.
	V2 clients.V2

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// New creates a new notification service client.
func New(v2 clients.V2, cache ...*core.Cache) *Client {
	var c *core.Cache
	if len(cache) > 0 {
		c = cache[0]
	} else {
		c = core.NewCache()
	}

	return &Client{
		V2:    v2,
		Cache: c,
	}
}

// Notification returns the notification with the given ID.
// After a successful fetch, the notification will be cached.
func (c *Client) Notification(id string, nmc *notification_repo.Client) (*model.Notification, error) {
	notification := c.Cache.GetNotification(id)
	if notification == nil {
		var err error
		notification, err = nmc.GetByID(id)
		if err != nil {
			return nil, err
		}

		c.Cache.StoreNotification(notification)
	}

	return notification, nil
}

// Notifications returns a list of notifications.
// After a successful fetch, the notifications will be cached.
func (c *Client) Notifications(nmc *notification_repo.Client) ([]model.Notification, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	notifications, err := nmc.List()
	if err != nil {
		return nil, err
	}

	for _, user := range notifications {
		c.Cache.StoreNotification(&user)
	}

	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].CreatedAt.After(notifications[j].CreatedAt)
	})

	return notifications, nil
}

// RefreshNotification clears the cache for the notification with the given ID and fetches it again.
// After a successful fetch, the notification will be cached.
func (c *Client) RefreshNotification(id string, umc *notification_repo.Client) (*model.Notification, error) {
	notification, err := umc.GetByID(id)
	if err != nil {
		return nil, err
	}

	c.Cache.StoreNotification(notification)
	return notification, nil
}
