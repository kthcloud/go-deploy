package status_codes

const (
	Unknown = 0

	ResourceCreating         = 10025
	ResourceDeleting         = 10026
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
	ResourceCrashLoop        = 10039
	ResourceMountFailed      = 10040
	ResourceImagePullFailed  = 10041

	JobPending    = 10140
	JobFinished   = 10141
	JobFailed     = 10142
	JobRunning    = 10143
	JobTerminated = 10144

	Success                  = 20001
	InvalidParams            = 20002
	Error                    = 20004
	ResourceValidationFailed = 20005
)

var MsgFlags = map[int]string{
	Unknown: "unknown",

	ResourceUnknown: "resourceUnknown",

	ResourceCreating:         "resourceCreating",
	ResourceDeleting:         "resourceDeleting",
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
	ResourceCrashLoop:        "resourceCrashLoop",
	ResourceMountFailed:      "resourceMountFailed",
	ResourceImagePullFailed:  "resourceImagePullFailed",

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
