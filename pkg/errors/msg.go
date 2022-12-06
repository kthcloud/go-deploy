package errors

var MsgFlags = map[int]string{
	SUCCESS:        "ok",
	ERROR:          "fail",
	INVALID_PARAMS: "invalid params",

	ERROR_PROJECT_BEING_CREATED: "project is currently being created",
	ERROR_PROJECT_BEING_DELETED: "project is currently being deleted",
	ERROR_PROJECT_EXISTS: "project name already exists",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
