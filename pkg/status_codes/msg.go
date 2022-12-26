package status_codes

var MsgFlags = map[int]string{
	Success:       "success",
	Error:         "error",
	InvalidParams: "invalidParams",

	ResourceBeingCreated:  "deploymentBeingCreated",
	ResourceBeingDeleted:  "deploymentBeingDeleted",
	ResourceAlreadyExists: "deploymentAlreadyExists",
	DeploymentCreated:     "deploymentCreated",
	DeploymentNotFound:    "deploymentNotFound",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[Error]
}
