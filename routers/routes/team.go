package routes

import (
	"go-deploy/routers/api/v1/v1_teams"
)

const (
	TeamsPath = "/v1/teams"
	TeamPath  = "/v1/teams/:teamId"
)

type TeamRoutingGroup struct{ RoutingGroupBase }

func TeamRoutes() *TeamRoutingGroup {
	return &TeamRoutingGroup{}
}

func (group TeamRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: TeamsPath, HandlerFunc: v1_teams.List},
		{Method: "GET", Pattern: TeamPath, HandlerFunc: v1_teams.Get},
		{Method: "POST", Pattern: TeamsPath, HandlerFunc: v1_teams.Create},
		{Method: "POST", Pattern: TeamPath, HandlerFunc: v1_teams.Update},
		{Method: "DELETE", Pattern: TeamPath, HandlerFunc: v1_teams.Delete},
	}
}
