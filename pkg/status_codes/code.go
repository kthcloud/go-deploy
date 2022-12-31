package status_codes

const (
	Unknown       = 0
	Success       = 200
	Error         = 500
	InvalidParams = 400

	ResourceUnknown = 10000

	ResourceCreated  = 10010
	ResourceNotFound = 10011

	ResourceBeingCreated = 10020
	ResourceBeingDeleted = 10021

	ResourceStarting = 10031
	ResourceRunning  = 10032
	ResourceStopping = 10033
	ResourceStopped  = 10034
	ResourceError    = 10035

	ResourceValidationFailed = 20001
	ResourceAlreadyExists    = 20002
)
