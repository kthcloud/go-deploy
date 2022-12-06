package errors

var MsgFlags = map[int]string{
	SUCCESS:        "SUCCESS",
	ERROR:          "ERROR",
	INVALID_PARAMS: "INVALID_PARAMS",

	ERROR_PROJECT_BEING_CREATED: "ERROR_PROJECT_BEING_CREATED",
	ERROR_PROJECT_BEING_DELETED: "ERROR_PROJECT_BEING_DELETED",
	ERROR_PROJECT_EXISTS:        "ERROR_PROJECT_EXISTS",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
