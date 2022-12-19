package models


type ApiError struct {
	Errors []struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"errors"`
}