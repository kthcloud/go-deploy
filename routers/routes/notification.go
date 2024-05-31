package routes

import v2 "go-deploy/routers/api/v2"

const (
	NotificationsPath = "/v2/notifications"
	NotificationPath  = "/v2/notifications/:notificationId"
)

type NotificationRoutingGroup struct{ RoutingGroupBase }

func NotificationRoutes() *NotificationRoutingGroup {
	return &NotificationRoutingGroup{}
}

func (group *NotificationRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: NotificationsPath, HandlerFunc: v2.ListNotifications},
		{Method: "GET", Pattern: NotificationPath, HandlerFunc: v2.GetNotification},
		{Method: "POST", Pattern: NotificationPath, HandlerFunc: v2.UpdateNotification},
		{Method: "DELETE", Pattern: NotificationPath, HandlerFunc: v2.DeleteNotification},
	}
}
