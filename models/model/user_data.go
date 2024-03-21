package model

import (
	"go-deploy/dto/v1/body"
	"time"
)

type UserData struct {
	ID     string `bson:"id"`
	Data   string `bson:"data"`
	UserID string `bson:"userId"`

	CreatedAt time.Time `bson:"createdAt"`
}

func (ud *UserData) ToDTO() body.UserDataRead {
	return body.UserDataRead{
		ID:     ud.ID,
		UserID: ud.UserID,
		Data:   ud.Data,
	}
}
