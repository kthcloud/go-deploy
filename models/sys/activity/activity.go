package activity

import "time"

type Activity struct {
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"createdAt"`
}
