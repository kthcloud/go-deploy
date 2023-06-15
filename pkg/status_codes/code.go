package status_codes

const (
	Unknown       = 0
	Success       = 200
	Error         = 500
	InvalidParams = 400

	ResourceUnknown = 10000

	ResourceCreated = 10010
	ResourceUpdated = 10011
	ResourceDeleted = 10012

	ResourceNotCreated   = 10020
	ResourceNotFound     = 10021
	ResourceNotUpdated   = 10022
	ResourceNotReady     = 10023
	ResourceNotAvailable = 10024
	ResourceBeingCreated = 10025
	ResourceBeingDeleted = 10026

	ResourceStarting   = 10031
	ResourceRunning    = 10032
	ResourceStopping   = 10033
	ResourceStopped    = 10034
	ResourceRestarting = 10035
	ResourceBuilding   = 10036
	ResourceError      = 10037

	JobPending  = 10040
	JobFinished = 10041
	JobFailed   = 10042
	JobRunning  = 10043

	ResourceValidationFailed = 20001
)
