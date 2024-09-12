package notifications

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/notification_repo"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/utils"
	"github.com/kthcloud/go-deploy/service/v2/notifications/opts"
)

// Get retrieves a notification by ID.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.Notification, error) {
	_ = utils.GetFirstOrDefault(opts)

	client := notification_repo.New()

	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		client.WithUserID(c.V2.Auth().User.ID)
	}

	return c.Notification(id, client)
}

// List retrieves a list of notifications.
func (c *Client) List(opts ...opts.ListOpts) ([]model.Notification, error) {
	o := utils.GetFirstOrDefault(opts)

	nmc := notification_repo.New()

	if o.Pagination != nil {
		nmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != nil {
		// Specific user's notifications are requested
		if !c.V2.HasAuth() || c.V2.Auth().User.ID == *o.UserID || c.V2.Auth().User.IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.V2.Auth().User.ID
		}
	} else {
		// All notifications are requested
		if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
			effectiveUserID = c.V2.Auth().User.ID
		}
	}

	if effectiveUserID != "" {
		nmc.WithUserID(effectiveUserID)
	}

	return c.Notifications(nmc)
}

// Create creates a new notification.
func (c *Client) Create(id, userID string, params *model.NotificationCreateParams) (*model.Notification, error) {
	return notification_repo.New().Create(id, userID, params)
}

// Update updates the notification with the given ID.
func (c *Client) Update(id string, dtoNotificationUpdate *body.NotificationUpdate) (*model.Notification, error) {
	nmc := notification_repo.New()

	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		nmc.WithUserID(c.V2.Auth().User.ID)
	}

	notification, err := c.Notification(id, nmc)
	if err != nil {
		return nil, err
	}

	if notification == nil {
		return nil, nil
	}

	if dtoNotificationUpdate.Read && notification.ReadAt.IsZero() {
		err = nmc.MarkReadByID(id)
		if err != nil {
			return nil, err
		}
	}

	if dtoNotificationUpdate.Toasted && notification.ToastedAt.IsZero() {
		err = nmc.MarkToastedByID(id)
		if err != nil {
			return nil, err
		}
	}

	return c.RefreshNotification(id, nmc)
}

// Delete deletes the notification with the given ID.
func (c *Client) Delete(id string) error {
	client := notification_repo.New()

	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		client.WithUserID(c.V2.Auth().User.ID)
	}

	exists, err := client.ExistsByID(id)
	if err != nil {
		return err
	}

	if !exists {
		return sErrors.NotificationNotFoundErr
	}

	return client.DeleteByID(id)
}
