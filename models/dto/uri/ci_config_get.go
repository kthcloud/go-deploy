package uri

type CiConfigGet struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}
