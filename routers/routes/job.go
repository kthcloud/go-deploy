package routes

import v1 "go-deploy/routers/api/v1"

const (
	JobsPath = "/v1/jobs"
	JobPath  = "/v1/jobs/:jobId"
)

type JobRoutingGroup struct{ RoutingGroupBase }

func JobRoutes() *JobRoutingGroup {
	return &JobRoutingGroup{}
}

func (group *JobRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: JobsPath, HandlerFunc: v1.ListJobs},
		{Method: "GET", Pattern: JobPath, HandlerFunc: v1.GetJob},
		{Method: "POST", Pattern: JobPath, HandlerFunc: v1.UpdateJob},
	}
}
