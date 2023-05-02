package uri

type LogsGet struct {
	DeploymentID string `uri:"deploymentId" bind:"required,uuid4"`
}
