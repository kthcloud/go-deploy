package user_data

import "time"

type UserData struct {
	ID     string `bson:"id"`
	Data   string `bson:"data"`
	UserID string `bson:"userId"`

	CreatedAt time.Time `bson:"createdAt"`
}
