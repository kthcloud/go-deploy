package uri

type DeploymentGet struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}
