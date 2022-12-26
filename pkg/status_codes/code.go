package status_codes

const (
	Success       = 200
	Error         = 500
	InvalidParams = 400

	DeploymentCreated    = 10001
	DeploymentNotFound   = 10002
	ResourceBeingCreated = 10003
	ResourceBeingDeleted = 10004

	ResourceValidationFailed = 20001
	ResourceAlreadyExists    = 20002
)
