package models

import "encoding/json"

type Response struct {
	// Code HTTP status code.
	Code int `json:"code,omitempty"`

	// Data API response data.
	Data json.RawMessage `json:"data,omitempty"`

	// Message API response message.
	Message string `json:"message,omitempty"`

	// Return API response/error code.
	Return int `json:"return,omitempty"`

	// Status HTTP status message.
	Status string `json:"status,omitempty"`
}
