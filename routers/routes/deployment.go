package routes

import (
	"github.com/gin-gonic/gin"
	"go-deploy/routers/api/v1/middleware"
	"go-deploy/routers/api/v1/v1_deployment"
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
		{Method: "GET", Pattern: DeploymentsPath, HandlerFunc: v1_deployment.List},

		{Method: "GET", Pattern: DeploymentPath, HandlerFunc: v1_deployment.Get},
		{Method: "POST", Pattern: DeploymentsPath, HandlerFunc: v1_deployment.Create, Middleware: []gin.HandlerFunc{middleware.CreateSM()}},
		{Method: "POST", Pattern: DeploymentPath, HandlerFunc: v1_deployment.Update},
		{Method: "DELETE", Pattern: DeploymentPath, HandlerFunc: v1_deployment.Delete},

		{Method: "GET", Pattern: DeploymentCiConfigPath, HandlerFunc: v1_deployment.GetCiConfig},
		{Method: "POST", Pattern: DeploymentCommandPath, HandlerFunc: v1_deployment.DoCommand},
		{Method: "GET", Pattern: DeploymentLogsPath, HandlerFunc: v1_deployment.GetLogsSSE, Middleware: []gin.HandlerFunc{middleware.SseSetup()}},
	}
}

func (group *DeploymentRoutingGroup) HookRoutes() []Route {
	return []Route{
		{Method: "POST", Pattern: DeploymentHarborHookPath, HandlerFunc: v1_deployment.HandleHarborHook},
	}
}
