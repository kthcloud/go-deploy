package models

type RecordPublic struct {
	ID         string   `bson:"ID"`
	Hostname   string   `bson:"hostname"`
	RecordType string   `bson:"recordType"`
	TTL        uint32   `bson:"ttl"`
	Content    []string `bson:"content"`
}
