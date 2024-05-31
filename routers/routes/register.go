package routes

import v2 "go-deploy/routers/api/v2"

const (
	RegisterPath = "/v2/register"
)

type RegisterRoutingGroup struct{ RoutingGroupBase }

func RegisterRoutes() *RegisterRoutingGroup {
	return &RegisterRoutingGroup{}
}

func (group *RegisterRoutingGroup) PublicRoutes() []Route {
	return []Route{
		{Method: "POST", Pattern: RegisterPath, HandlerFunc: v2.Register},
	}
}
