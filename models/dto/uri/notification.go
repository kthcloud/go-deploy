package uri

type NotificationGet struct {
	NotificationID string `uri:"notificationId" binding:"required,uuid4"`
}

type NotificationUpdate struct {
	NotificationID string `uri:"notificationId" binding:"required,uuid4"`
}

type NotificationDelete struct {
	NotificationID string `uri:"notificationId" binding:"required,uuid4"`
}
