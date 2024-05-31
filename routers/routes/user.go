package routes

import v2 "go-deploy/routers/api/v2"

const (
	UsersPath  = "/v2/users"
	UserPath   = "/v2/users/:userId"
	ApiKeyPath = "/v2/users/:userId/apiKeys"
)

type UserRoutingGroup struct{ RoutingGroupBase }

func UserRoutes() *UserRoutingGroup {
	return &UserRoutingGroup{}
}

func (group *UserRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: UsersPath, HandlerFunc: v2.ListUsers},
		{Method: "GET", Pattern: UserPath, HandlerFunc: v2.GetUser},
		{Method: "POST", Pattern: UsersPath, HandlerFunc: v2.UpdateUser}, // update using id in the token
		{Method: "POST", Pattern: UserPath, HandlerFunc: v2.UpdateUser},  // update using id in the path

		{Method: "POST", Pattern: ApiKeyPath, HandlerFunc: v2.CreateApiKey},
	}
}
