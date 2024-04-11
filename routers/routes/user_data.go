package routes

import v1 "go-deploy/routers/api/v1"

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
		{Method: "GET", Pattern: UserDataRootPath, HandlerFunc: v1.ListUserData},
		{Method: "GET", Pattern: UserDataPath, HandlerFunc: v1.GetUserData},
		{Method: "POST", Pattern: UserDataRootPath, HandlerFunc: v1.CreateUserData},
		{Method: "POST", Pattern: UserDataPath, HandlerFunc: v1.UpdateUserData},
		{Method: "DELETE", Pattern: UserDataPath, HandlerFunc: v1.DeleteUserData},
	}
}
