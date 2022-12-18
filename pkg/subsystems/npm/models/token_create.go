package models

type TokenCreate struct {
	Identity string `json:"identity,omitempty"`
	Secret   string `json:"secret,omitempty"`
}