package uri

type DeploymentDelete struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}
