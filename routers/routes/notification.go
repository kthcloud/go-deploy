package routes

import "go-deploy/routers/api/v1/v1_notification"

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
		{Method: "GET", Pattern: NotificationsPath, HandlerFunc: v1_notification.List},
		{Method: "GET", Pattern: NotificationPath, HandlerFunc: v1_notification.Get},
		{Method: "POST", Pattern: NotificationPath, HandlerFunc: v1_notification.Update},
		{Method: "DELETE", Pattern: NotificationPath, HandlerFunc: v1_notification.Delete},
	}
}
