package status_codes

const (
	Success       = 200
	Error         = 500
	InvalidParams = 400

	DeploymentCreated      = 10001
	DeploymentNotFound     = 10002
	DeploymentBeingCreated = 10003
	DeploymentBeingDeleted = 10004

	DeploymentValidationFailed = 20001
	DeploymentAlreadyExists    = 20002
)
