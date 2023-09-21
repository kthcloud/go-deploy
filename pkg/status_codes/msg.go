package status_codes

var MsgFlags = map[int]string{
	Unknown:       "unknown",
	Success:       "success",
	Error:         "error",
	InvalidParams: "invalidParams",

	ResourceUnknown: "resourceUnknown",

	ResourceCreated: "resourceCreated",
	ResourceUpdated: "resourceUpdated",

	ResourceNotCreated:       "resourceNotCreated",
	ResourceNotFound:         "resourceNotFound",
	ResourceNotUpdated:       "resourceNotModified",
	ResourceNotReady:         "resourceNotReady",
	ResourceNotAvailable:     "resourceNotAvailable",
	ResourceBeingCreated:     "resourceBeingCreated",
	ResourceBeingDeleted:     "resourceBeingDeleted",
	ResourceCreatingSnapshot: "resourceCreatingSnapshot",
	ResourceApplyingSnapshot: "resourceApplyingSnapshot",

	ResourceStarting:   "resourceStarting",
	ResourceRunning:    "resourceRunning",
	ResourceStopping:   "resourceStopping",
	ResourceStopped:    "resourceStopped",
	ResourceRestarting: "resourceRestarting",
	ResourceBuilding:   "resourceBuilding",
	ResourceError:      "resourceError",

	JobPending:    "pending",
	JobRunning:    "running",
	JobFailed:     "failed",
	JobFinished:   "finished",
	JobTerminated: "terminated",

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
