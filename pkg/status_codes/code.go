package status_codes

const (
	Success       = 200
	Error         = 500
	InvalidParams = 400

	ProjectCreated      = 10001
	ProjectNotFound     = 10002
	ProjectBeingCreated = 10003
	ProjectBeingDeleted = 10004

	ProjectValidationFailed = 20001
	ProjectAlreadyExists    = 20002
)
