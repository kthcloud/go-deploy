package uri

type DeploymentGet struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}

type DeploymentDelete struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}

type DeploymentUpdate struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}

type CiConfigGet struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}

type DeploymentCommand struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}

type LogsGet struct {
	DeploymentID string `uri:"deploymentId" bind:"required,uuid4"`
}

type BuildGet struct {
	DeploymentID string `uri:"deploymentId" binding:"required,uuid4"`
}

type SmGet struct {
	SmID string `uri:"storageManagerId" binding:"required,uuid4"`
}

type SmDelete struct {
	SmID string `uri:"storageManagerId" binding:"required,uuid4"`
}
