package routes

import v1 "go-deploy/routers/api/v1"

const (
	NotificationsPath = "/v1/notifications"
	NotificationPath  = "/v1/notifications/:notificationId"
)

type NotificationRoutingGroup struct{ RoutingGroupBase }

func NotificationRoutes() *NotificationRoutingGroup {
	return &NotificationRoutingGroup{}
}

func (group *NotificationRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: NotificationsPath, HandlerFunc: v1.ListNotifications},
		{Method: "GET", Pattern: NotificationPath, HandlerFunc: v1.GetNotification},
		{Method: "POST", Pattern: NotificationPath, HandlerFunc: v1.UpdateNotification},
		{Method: "DELETE", Pattern: NotificationPath, HandlerFunc: v1.DeleteNotification},
	}
}
