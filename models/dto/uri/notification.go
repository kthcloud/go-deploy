package uri

type NotificationUpdate struct {
	NotificationID string `uri:"notificationId" binding:"required,uuid4"`
}

type NotificationDelete struct {
	NotificationID string `uri:"notificationId" binding:"required,uuid4"`
}
