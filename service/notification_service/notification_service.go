package notification_service

import (
	"go-deploy/models/dto/body"
	notificationModels "go-deploy/models/sys/notification"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
)

// Get retrieves a notification by ID.
func (c *Client) Get(id string, opts ...GetOpts) (*notificationModels.Notification, error) {
	_ = service.GetFirstOrDefault(opts)

	client := notificationModels.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		client.WithUserID(c.Auth.UserID)
	}

	return c.Notification(id, client)
}

// List retrieves a list of notifications.
func (c *Client) List(opts ...ListOpts) ([]notificationModels.Notification, error) {
	o := service.GetFirstOrDefault(opts)

	nmc := notificationModels.New()

	if o.Pagination != nil {
		nmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != nil {
		// Specific user's notifications are requested
		if c.Auth == nil || c.Auth.UserID == *o.UserID || c.Auth.IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.Auth.UserID
		}
	} else {
		// All notifications are requested
		if c.Auth != nil && !c.Auth.IsAdmin {
			effectiveUserID = c.Auth.UserID
		}
	}

	if effectiveUserID != "" {
		nmc.WithUserID(effectiveUserID)
	}

	return c.Notifications(nmc)
}

// Create creates a new notification.
func (c *Client) Create(id, userID string, params *notificationModels.CreateParams) (*notificationModels.Notification, error) {
	return notificationModels.New().Create(id, userID, params)
}

// Update updates the notification with the given ID.
func (c *Client) Update(id string, dtoNotificationUpdate *body.NotificationUpdate) (*notificationModels.Notification, error) {
	nmc := notificationModels.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		nmc.WithUserID(c.Auth.UserID)
	}

	notification, err := c.Notification(id, nmc)
	if err != nil {
		return nil, err
	}

	if notification == nil {
		return nil, nil
	}

	if dtoNotificationUpdate.Read && !notification.ReadAt.IsZero() {
		err = nmc.MarkReadByID(id)
		if err != nil {
			return nil, err
		}
	}

	return c.RefreshNotification(id, nmc)
}

// Delete deletes the notification with the given ID.
func (c *Client) Delete(id string) error {
	client := notificationModels.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		client.WithUserID(c.Auth.UserID)
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
