package models

type UserHttpEvent []struct {
	UserID string
	Event  string
	IP     string
}
