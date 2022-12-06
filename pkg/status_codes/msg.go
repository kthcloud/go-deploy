package status_codes

var MsgFlags = map[int]string{
	Success:       "Success",
	Error:         "Error",
	InvalidParams: "InvalidParams",

	ProjectBeingCreated:  "ProjectBeingCreated",
	ProjectBeingDeleted:  "ProjectBeingDeleted",
	ProjectAlreadyExists: "ProjectAlreadyExists",
	ProjectCreated:       "ProjectCreated",
	ProjectNotFound:      "ProjectNotFound",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[Error]
}
