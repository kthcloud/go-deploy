package routes

import (
	"github.com/gin-gonic/gin"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/routers/api/v1/middleware"
)

const (
	DeploymentsPath          = "/v1/deployments"
	DeploymentPath           = "/v1/deployments/:deploymentId"
	DeploymentCiConfigPath   = "/v1/deployments/:deploymentId/ciConfig"
	DeploymentCommandPath    = "/v1/deployments/:deploymentId/command"
	DeploymentLogsPath       = "/v1/deployments/:deploymentId/logs-sse"
	DeploymentHarborHookPath = "/v1/hooks/deployments/harbor"
)

type DeploymentRoutingGroup struct{ RoutingGroupBase }

func DeploymentRoutes() *DeploymentRoutingGroup { return &DeploymentRoutingGroup{} }

func (group *DeploymentRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: DeploymentsPath, HandlerFunc: v1.ListDeployments},

		{Method: "GET", Pattern: DeploymentPath, HandlerFunc: v1.GetDeployment},
		{Method: "POST", Pattern: DeploymentsPath, HandlerFunc: v1.CreateDeployment,
			Middleware: []gin.HandlerFunc{middleware.CreateSM()}},
		{Method: "POST", Pattern: DeploymentPath, HandlerFunc: v1.UpdateDeployment},
		{Method: "DELETE", Pattern: DeploymentPath, HandlerFunc: v1.DeleteDeployment},

		{Method: "GET", Pattern: DeploymentCiConfigPath, HandlerFunc: v1.GetCiConfig},
		{Method: "POST", Pattern: DeploymentCommandPath, HandlerFunc: v1.DoDeploymentCommand},
		{Method: "GET", Pattern: DeploymentLogsPath, HandlerFunc: v1.GetLogs, Middleware: []gin.HandlerFunc{middleware.SseSetup()}},
	}
}

func (group *DeploymentRoutingGroup) HookRoutes() []Route {
	return []Route{
		{Method: "POST", Pattern: DeploymentHarborHookPath, HandlerFunc: v1.HandleHarborHook},
	}
}
