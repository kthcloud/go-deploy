package notification

import "go-deploy/models/dto/body"

func (notification *Notification) ToDTO() body.NotificationRead {
	return body.NotificationRead{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Content:   notification.Content,
		CreatedAt: notification.CreatedAt,
		ReadAt:    notification.ReadAt,
	}
}
