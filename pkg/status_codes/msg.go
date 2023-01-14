package status_codes

var MsgFlags = map[int]string{
	Unknown:       "unknown",
	Success:       "success",
	Error:         "error",
	InvalidParams: "invalidParams",

	ResourceUnknown: "resourceUnknown",

	ResourceCreated:  "resourceCreated",
	ResourceNotFound: "resourceNotFound",

	ResourceBeingCreated: "resourceBeingCreated",
	ResourceBeingDeleted: "resourceBeingDeleted",

	ResourceStarting: "resourceStarting",
	ResourceRunning:  "resourceRunning",
	ResourceStopping: "resourceStopping",
	ResourceStopped:  "resourceStopped",
	ResourceError:    "resourceError",

	ResourceValidationFailed: "resourceValidationFailed",
	ResourceAlreadyExists:    "resourceAlreadyExists",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[Error]
}
