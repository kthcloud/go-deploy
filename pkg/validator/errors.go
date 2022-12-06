package validator

import "errors"

var (
	errStringToInt          = errors.New("validator: unable to parse string to integer")
	errStringToFloat        = errors.New("validator: unable to parse string to float")
	errRequireRules         = errors.New("validator: provide at least rules for Validate* method")
	errValidateArgsMismatch = errors.New("validator: provide at least *http.Request and rules for Validate method")
	errInvalidArgument      = errors.New("validator: invalid number of argument")
	errRequirePtr           = errors.New("validator: provide pointer to the data structure")
	errRequireData          = errors.New("validator: provide non-nil data structure for ValidateStruct method")
	errRequestNotAccepted   = errors.New("validator: cannot provide an *http.Request for ValidateStruct method")
)
