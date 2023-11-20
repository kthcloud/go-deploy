package notification_service

import (
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	notificationModel "go-deploy/models/sys/notification"
	"go-deploy/service"
)

func CreateNotification(id, userID string, params *notificationModel.CreateParams) error {
	_, err := notificationModel.New().Create(id, userID, params)
	return err
}

func GetByIdWithAuth(id string, auth *service.AuthInfo) (*notificationModel.Notification, error) {
	client := notificationModel.New()

	if !auth.IsAdmin {
		client.RestrictToUserID(auth.UserID)
	}

	return client.GetByID(id)
}

func ListAuth(allUsers bool, userID *string, auth *service.AuthInfo, pagination *query.Pagination) ([]notificationModel.Notification, error) {
	client := notificationModel.New()

	if pagination != nil {
		client.AddPagination(pagination.Page, pagination.PageSize)
	}

	if userID != nil {
		if *userID != auth.UserID && !auth.IsAdmin {
			return nil, nil
		}
		client.RestrictToUserID(*userID)
	} else if !allUsers || (allUsers && !auth.IsAdmin) {
		client.RestrictToUserID(auth.UserID)
	}

	return client.List()
}

func UpdateAuth(id string, dtoNotificationUpdate *body.NotificationUpdate, auth *service.AuthInfo) (*notificationModel.Notification, error) {
	client := notificationModel.New()

	if !auth.IsAdmin {
		client.RestrictToUserID(auth.UserID)
	}

	notification, err := client.GetByID(id)
	if err != nil {
		return nil, err
	}

	params := &notificationModel.UpdateParams{}
	params.FromDTO(dtoNotificationUpdate)

	// if the notification is already read, we don't want to update it to a newer read time
	// the user should unread it first
	if notification.ReadAt != nil && params.ReadAt != nil {
		params.ReadAt = nil
	}

	err = client.UpdateWithParamsByID(id, params)
	if err != nil {
		return nil, err
	}

	return client.GetByID(id)
}

func DeleteAuth(id string, auth *service.AuthInfo) error {
	client := notificationModel.New()
	if !auth.IsAdmin {
		client.RestrictToUserID(auth.UserID)
	}
	return client.DeleteByID(id)
}
