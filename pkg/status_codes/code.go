package status_codes

const (
	Success       = 200
	Error         = 500
	InvalidParams = 400

	ResourceCreated  = 10001
	ResourceNotFound = 10002
	ResourceRunning  = 10003
	ResourceBeingCreated = 10004
	ResourceBeingDeleted = 10005

	ResourceValidationFailed = 20001
	ResourceAlreadyExists    = 20002
)
