package body

import "time"

type ApiKeyCreate struct {
	Name      string    `json:"name" binding:"required"`
	ExpiresAt time.Time `json:"expiresAt" binding:"required,time_in_future"`
}

type ApiKeyCreated struct {
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}
