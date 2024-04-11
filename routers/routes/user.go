package routes

import v1 "go-deploy/routers/api/v1"

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
		{Method: "GET", Pattern: UsersPath, HandlerFunc: v1.ListUsers},
		{Method: "GET", Pattern: UserPath, HandlerFunc: v1.GetUser},
		{Method: "POST", Pattern: UsersPath, HandlerFunc: v1.UpdateUser}, // update using id in the token
		{Method: "POST", Pattern: UserPath, HandlerFunc: v1.UpdateUser},  // update using id in the path
	}
}
