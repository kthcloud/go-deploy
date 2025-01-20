package routes

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/kthcloud/go-deploy/routers/api/v2"
	"github.com/kthcloud/go-deploy/routers/api/v2/middleware"
)

const (
	DeploymentsPath          = "/v2/deployments"
	DeploymentPath           = "/v2/deployments/:deploymentId"
	DeploymentCiConfigPath   = "/v2/deployments/:deploymentId/ciConfig"
	DeploymentCommandPath    = "/v2/deployments/:deploymentId/command"
	DeploymentLogsPath       = "/v2/deployments/:deploymentId/logs-sse"
	DeploymentHarborHookPath = "/v2/hooks/deployments/harbor"
)

type DeploymentRoutingGroup struct{ RoutingGroupBase }

func DeploymentRoutes() *DeploymentRoutingGroup { return &DeploymentRoutingGroup{} }

func (group *DeploymentRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: DeploymentsPath, HandlerFunc: v2.ListDeployments},

		{Method: "GET", Pattern: DeploymentPath, HandlerFunc: v2.GetDeployment},
		{Method: "POST", Pattern: DeploymentsPath, HandlerFunc: v2.CreateDeployment, Middleware: []gin.HandlerFunc{middleware.CreateSM()}},
		{Method: "POST", Pattern: DeploymentPath, HandlerFunc: v2.UpdateDeployment},
		{Method: "DELETE", Pattern: DeploymentPath, HandlerFunc: v2.DeleteDeployment},

		{Method: "GET", Pattern: DeploymentCiConfigPath, HandlerFunc: v2.GetCiConfig},
		{Method: "POST", Pattern: DeploymentCommandPath, HandlerFunc: v2.DoDeploymentCommand},
		{Method: "GET", Pattern: DeploymentLogsPath, HandlerFunc: v2.GetLogs, Middleware: []gin.HandlerFunc{middleware.SseSetup()}},
	}
}

func (group *DeploymentRoutingGroup) HookRoutes() []Route {
	return []Route{
		{Method: "POST", Pattern: DeploymentHarborHookPath, HandlerFunc: v2.HandleHarborHook},
	}
}
