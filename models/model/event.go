package model

import "time"

const (
	// TypeHttpRequest is the event type of HTTP requests.
	TypeHttpRequest = "httpRequest"
)

type Source struct {
	UserID *string `bson:"userId,omitempty"`
	IP     *string `bson:"ip"`
}

type Event struct {
	ID        string    `bson:"id"`
	Type      string    `bson:"type"`
	CreatedAt time.Time `bson:"createdAt"`

	// Source is the source of an event, meaning from where the event originated.
	Source *Source `bson:"source,omitempty"`

	// Metadata contains any data related to the event.
	Metadata map[string]interface{} `bson:"metadata,omitempty"`
}
