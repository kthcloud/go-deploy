package models

type ProjectPublic struct {
	ID        int64  `bson:"id"`
	Name      string `bson:"name"`
	ImportURL string `bson:"importUrl"`
}
