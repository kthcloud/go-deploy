package notification

import "go-deploy/models/dto/body"

func (notification *Notification) ToDTO() body.NotificationRead {
	return body.NotificationRead{
		ID:      notification.ID,
		Type:    notification.Type,
		Content: notification.Content,
		ReadAt:  notification.ReadAt,
	}
}
