package dto

type JobRead struct {
	ID     string `bson:"id" json:"id"`
	UserID string `bson:"userId" json:"userId"`
	Type   string `bson:"type" json:"type"`
	Status string `bson:"status" json:"status"`
}
