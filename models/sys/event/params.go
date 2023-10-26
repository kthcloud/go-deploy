package event

type CreateParams struct {
	Type     string                 `bson:"type"`
	Source   *Source                `bson:"source,omitempty"`
	Metadata map[string]interface{} `bson:"metadata"`
}
