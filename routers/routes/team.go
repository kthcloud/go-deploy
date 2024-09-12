package routes

import v2 "github.com/kthcloud/go-deploy/routers/api/v2"

const (
	TeamsPath = "/v2/teams"
	TeamPath  = "/v2/teams/:teamId"
)

type TeamRoutingGroup struct{ RoutingGroupBase }

func TeamRoutes() *TeamRoutingGroup {
	return &TeamRoutingGroup{}
}

func (group TeamRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: TeamsPath, HandlerFunc: v2.ListTeams},
		{Method: "GET", Pattern: TeamPath, HandlerFunc: v2.GetTeam},
		{Method: "POST", Pattern: TeamsPath, HandlerFunc: v2.CreateTeam},
		{Method: "POST", Pattern: TeamPath, HandlerFunc: v2.UpdateTeam},
		{Method: "DELETE", Pattern: TeamPath, HandlerFunc: v2.DeleteTeam},
	}
}
