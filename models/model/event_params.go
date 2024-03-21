package model

type EventCreateParams struct {
	Type     string                 `bson:"type"`
	Source   *Source                `bson:"source,omitempty"`
	Metadata map[string]interface{} `bson:"metadata"`
}
