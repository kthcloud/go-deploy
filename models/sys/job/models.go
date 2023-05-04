package job

import "time"

type Job struct {
	ID        string                 `bson:"id" json:"id"`
	UserID    string                 `bson:"userId" json:"userId"`
	Type      string                 `bson:"type" json:"type"`
	Args      map[string]interface{} `bson:"args" json:"args"`
	CreatedAt time.Time              `bson:"createdAt" json:"createdAt"`
	Status    string                 `bson:"status" json:"status"`
	ErrorLogs []string               `bson:"errorLogs" json:"errorLogs"`
}

