package event

import "time"

const (
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

	Source   *Source                `bson:"source,omitempty"`
	Metadata map[string]interface{} `bson:"metadata,omitempty"`
}
