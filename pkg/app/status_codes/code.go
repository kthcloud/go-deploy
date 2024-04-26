package status_codes

const (
	Unknown = 0

	ResourceNotCreated       = 10020
	ResourceNotFound         = 10021
	ResourceNotUpdated       = 10022
	ResourceNotReady         = 10023
	ResourceNotAvailable     = 10024
	ResourceBeingCreated     = 10025
	ResourceBeingDeleted     = 10026
	ResourceCreatingSnapshot = 10027
	ResourceProvisioning     = 10028
	ResourceMigrating        = 10029
	ResourceUnknown          = 10030
	ResourceStarting         = 10031
	ResourceRunning          = 10032
	ResourceStopping         = 10033
	ResourceStopped          = 10034
	ResourceRestarting       = 10035
	ResourceBuilding         = 10036
	ResourceError            = 10037
	ResourceScaling          = 10038

	JobPending    = 10040
	JobFinished   = 10041
	JobFailed     = 10042
	JobRunning    = 10043
	JobTerminated = 10044

	Success                  = 20001
	InvalidParams            = 20002
	Error                    = 20004
	ResourceValidationFailed = 20005
)

var MsgFlags = map[int]string{
	Unknown: "unknown",

	ResourceUnknown: "resourceUnknown",

	ResourceNotCreated:       "resourceNotCreated",
	ResourceNotFound:         "resourceNotFound",
	ResourceNotUpdated:       "resourceNotModified",
	ResourceNotReady:         "resourceNotReady",
	ResourceNotAvailable:     "resourceNotAvailable",
	ResourceBeingCreated:     "resourceBeingCreated",
	ResourceBeingDeleted:     "resourceBeingDeleted",
	ResourceCreatingSnapshot: "resourceCreatingSnapshot",
	ResourceMigrating:        "resourceMigrating",
	ResourceProvisioning:     "resourceProvisioning",
	ResourceStarting:         "resourceStarting",
	ResourceRunning:          "resourceRunning",
	ResourceStopping:         "resourceStopping",
	ResourceStopped:          "resourceStopped",
	ResourceRestarting:       "resourceRestarting",
	ResourceBuilding:         "resourceBuilding",
	ResourceError:            "resourceError",
	ResourceScaling:          "resourceScaling",

	JobPending:    "pending",
	JobRunning:    "running",
	JobFailed:     "failed",
	JobFinished:   "finished",
	JobTerminated: "terminated",

	Success:                  "success",
	Error:                    "error",
	InvalidParams:            "invalidParams",
	ResourceValidationFailed: "resourceValidationFailed",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[Error]
}
