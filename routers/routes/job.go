package routes

import "go-deploy/routers/api/v1/v1_job"

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
		{Method: "GET", Pattern: JobsPath, HandlerFunc: v1_job.List},
		{Method: "GET", Pattern: JobPath, HandlerFunc: v1_job.Get},
		{Method: "POST", Pattern: JobPath, HandlerFunc: v1_job.Update},
	}
}
