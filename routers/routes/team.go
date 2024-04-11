package routes

import v1 "go-deploy/routers/api/v1"

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
		{Method: "GET", Pattern: TeamsPath, HandlerFunc: v1.ListTeams},
		{Method: "GET", Pattern: TeamPath, HandlerFunc: v1.GetTeam},
		{Method: "POST", Pattern: TeamsPath, HandlerFunc: v1.CreateTeam},
		{Method: "POST", Pattern: TeamPath, HandlerFunc: v1.UpdateTeam},
		{Method: "DELETE", Pattern: TeamPath, HandlerFunc: v1.DeleteTeam},
	}
}
