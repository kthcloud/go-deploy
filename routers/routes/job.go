package routes

import v2 "github.com/kthcloud/go-deploy/routers/api/v2"

const (
	JobsPath = "/v2/jobs"
	JobPath  = "/v2/jobs/:jobId"
)

type JobRoutingGroup struct{ RoutingGroupBase }

func JobRoutes() *JobRoutingGroup {
	return &JobRoutingGroup{}
}

func (group *JobRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: JobsPath, HandlerFunc: v2.ListJobs},
		{Method: "GET", Pattern: JobPath, HandlerFunc: v2.GetJob},
		{Method: "POST", Pattern: JobPath, HandlerFunc: v2.UpdateJob},
	}
}
