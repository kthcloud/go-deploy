package routes

import "go-deploy/routers/api/v1/v1_user"

const (
	UsersPath = "/v1/users"
	UserPath  = "/v1/users/:userId"
)

type UserRoutingGroup struct{ RoutingGroupBase }

func UserRoutes() *UserRoutingGroup {
	return &UserRoutingGroup{}
}

func (group *UserRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: UsersPath, HandlerFunc: v1_user.ListUsers},
		{Method: "GET", Pattern: UserPath, HandlerFunc: v1_user.Get},
		{Method: "POST", Pattern: UsersPath, HandlerFunc: v1_user.Update}, // update using id in the token
		{Method: "POST", Pattern: UserPath, HandlerFunc: v1_user.Update},  // update using id in the path
	}
}
