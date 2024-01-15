package models

// ApiError is the generic error returned by the Harbor API.
type ApiError struct {
	Errors []struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"errors"`
}
