package routes

import "go-deploy/routers/api/v1/v1_user"

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
		{Method: "GET", Pattern: TeamsPath, HandlerFunc: v1_user.ListTeams},
		{Method: "GET", Pattern: TeamPath, HandlerFunc: v1_user.GetTeam},
		{Method: "POST", Pattern: TeamsPath, HandlerFunc: v1_user.CreateTeam},
		{Method: "POST", Pattern: TeamPath, HandlerFunc: v1_user.UpdateTeam},
		{Method: "DELETE", Pattern: TeamPath, HandlerFunc: v1_user.DeleteTeam},
	}
}
