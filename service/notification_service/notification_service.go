package notification_service

import (
	"go-deploy/models/dto/body"
	notificationModel "go-deploy/models/sys/notification"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
)

func (c *Client) Get(id string, opts ...GetOpts) (*notificationModel.Notification, error) {
	_ = service.GetFirstOrDefault(opts)

	client := notificationModel.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		client.RestrictToUserID(c.Auth.UserID)
	}

	return c.Notification(id, client)
}

func (c *Client) List(opts ...ListOpts) ([]notificationModel.Notification, error) {
	o := service.GetFirstOrDefault(opts)

	nmc := notificationModel.New()

	if o.Pagination != nil {
		nmc.AddPagination(o.Pagination.Page, o.Pagination.PageSize)
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
		nmc.RestrictToUserID(effectiveUserID)
	}

	return c.Notifications(nmc)
}

func (c *Client) Create(id, userID string, params *notificationModel.CreateParams) (*notificationModel.Notification, error) {
	return notificationModel.New().Create(id, userID, params)
}

func (c *Client) Update(id string, dtoNotificationUpdate *body.NotificationUpdate) (*notificationModel.Notification, error) {
	nmc := notificationModel.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		nmc.RestrictToUserID(c.Auth.UserID)
	}

	notification, err := c.Notification(id, nmc)
	if err != nil {
		return nil, err
	}

	if notification == nil {
		return nil, nil
	}

	params := &notificationModel.UpdateParams{}
	params.FromDTO(dtoNotificationUpdate)

	// if the notification is already read, we don't want to update it to a newer read time
	// the user should unread it first
	if notification.ReadAt != nil && params.ReadAt != nil {
		params.ReadAt = nil
	}

	err = nmc.UpdateWithParamsByID(id, params)
	if err != nil {
		return nil, err
	}

	return c.RefreshNotification(id, nmc)
}

func (c *Client) Delete(id string) error {
	client := notificationModel.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		client.RestrictToUserID(c.Auth.UserID)
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
