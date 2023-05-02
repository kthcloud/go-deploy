package uri

type DeploymentGet struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}

type DeploymentDelete struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}

type CiConfigGet struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}

type DoCommand struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

type LogsGet struct {
	DeploymentID string `uri:"deploymentId" bind:"required,uuid4"`
}
