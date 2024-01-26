package notifications

import (
	notificationModels "go-deploy/models/sys/notification"
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"sort"
)

// Client is the client for the notification service.
type Client struct {
	// V1 is a reference to the parent client.
	V1 clients.V1

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// New creates a new notification service client.
func New(v1 clients.V1, cache ...*core.Cache) *Client {
	var c *core.Cache
	if len(cache) > 0 {
		c = cache[0]
	} else {
		c = core.NewCache()
	}

	return &Client{
		V1:    v1,
		Cache: c,
	}
}

// Notification returns the notification with the given ID.
// After a successful fetch, the notification will be cached.
func (c *Client) Notification(id string, nmc *notificationModels.Client) (*notificationModels.Notification, error) {
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
func (c *Client) Notifications(nmc *notificationModels.Client) ([]notificationModels.Notification, error) {
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
func (c *Client) RefreshNotification(id string, umc *notificationModels.Client) (*notificationModels.Notification, error) {
	notification, err := umc.GetByID(id)
	if err != nil {
		return nil, err
	}

	c.Cache.StoreNotification(notification)
	return notification, nil
}
