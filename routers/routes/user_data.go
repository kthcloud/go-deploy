package routes

import (
	"go-deploy/routers/api/v1/v1_user_data"
)

const (
	UserDataRootPath = "/v1/userData"
	UserDataPath     = "/v1/userData/:id"
)

type UserDataRoutingGroup struct{ RoutingGroupBase }

func UserDataRoutes() *UserDataRoutingGroup {
	return &UserDataRoutingGroup{}
}

func (group *UserDataRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: UserDataRootPath, HandlerFunc: v1_user_data.List},
		{Method: "GET", Pattern: UserDataPath, HandlerFunc: v1_user_data.Get},
		{Method: "POST", Pattern: UserDataRootPath, HandlerFunc: v1_user_data.Create},
		{Method: "POST", Pattern: UserDataPath, HandlerFunc: v1_user_data.Update},
		{Method: "DELETE", Pattern: UserDataPath, HandlerFunc: v1_user_data.Delete},
	}
}
