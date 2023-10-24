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

func GetByIdAuth(id string, auth *service.AuthInfo) (*notificationModel.Notification, error) {
	client := notificationModel.New()

	if !auth.IsAdmin {
		client.RestrictToUser(&auth.UserID)
	}

	return client.GetByID(id)
}

func GetManyAuth(allUsers bool, userID *string, auth *service.AuthInfo, pagination *query.Pagination) ([]notificationModel.Notification, error) {
	client := notificationModel.New()

	if pagination != nil {
		client.AddPagination(pagination.Page, pagination.PageSize)
	}

	if userID != nil {
		if *userID != auth.UserID && !auth.IsAdmin {
			return nil, nil
		}
		client.RestrictToUser(userID)
	} else if !allUsers || (allUsers && !auth.IsAdmin) {
		client.RestrictToUser(&auth.UserID)
	}

	return client.GetAll()
}

func UpdateAuth(id string, dtoNotificationUpdate *body.NotificationUpdate, auth *service.AuthInfo) error {
	client := notificationModel.New()

	if !auth.IsAdmin {
		client.RestrictToUser(&auth.UserID)
	}

	params := &notificationModel.UpdateParams{}
	params.FromDTO(dtoNotificationUpdate)

	return client.UpdateWithParamsByID(id, params)
}

func DeleteAuth(id string, auth *service.AuthInfo) error {
	client := notificationModel.New()
	if !auth.IsAdmin {
		client.RestrictToUser(&auth.UserID)
	}
	return client.DeleteByID(id)
}
