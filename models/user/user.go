package user

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/dto"
	"go-deploy/pkg/conf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type User struct {
	ID              string   `json:"id" bson:"id"`
	Username        string   `json:"username" bson:"username"`
	Email           string   `json:"email" bson:"email"`
	VmQuota         int      `json:"vmQuota" bson:"vmQuota"`
	DeploymentQuota int      `json:"deploymentQuota" bson:"deploymentQuota"`
	IsAdmin         bool     `json:"isAdmin" bson:"isAdmin"`
	IsPowerUser     bool     `json:"isPowerUser" bson:"isPowerUser"`
	PublicKeys      []string `json:"publicKeys" bson:"publicKeys"`
}

func (u *User) ToDTO() dto.UserRead {
	userRead := dto.UserRead{
		ID:              u.ID,
		Username:        u.Username,
		Email:           u.Email,
		Admin:           u.IsAdmin,
		VmQuota:         u.VmQuota,
		DeploymentQuota: u.DeploymentQuota,
		PowerUser:       u.IsPowerUser,
		PublicKeys:      u.PublicKeys,
	}

	if userRead.PublicKeys == nil {
		userRead.PublicKeys = []string{}
	}

	return userRead
}

func Create(id, username string) error {
	current, err := GetByID(id)
	if err != nil {
		return err
	}

	if current != nil {
		return nil
	}

	_, err = models.UserCollection.InsertOne(context.TODO(), User{
		ID:              id,
		Username:        username,
		Email:           "",
		VmQuota:         conf.Env.VM.DefaultQuota,
		DeploymentQuota: conf.Env.App.DefaultQuota,
		IsAdmin:         false,
		IsPowerUser:     false,
		PublicKeys:      []string{},
	})

	if err != nil {
		return fmt.Errorf("failed to create user info for %s. details: %s", username, err)
	}

	return nil
}

func GetByID(id string) (*User, error) {
	var userInfo User
	filter := bson.D{{"id", id}}
	err := models.UserCollection.FindOne(context.TODO(), filter).Decode(&userInfo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch user info by id %s. details: %s", id, err)
		return nil, err
	}

	return &userInfo, err
}

func GetAll() ([]User, error) {
	var users []User
	cursor, err := models.UserCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func Update(userID string, update *UserUpdate) error {
	updateData := bson.M{}

	addIfNotNil(updateData, "username", update.Username)
	addIfNotNil(updateData, "email", update.Email)
	addIfNotNil(updateData, "vmQuota", update.VmQuota)
	addIfNotNil(updateData, "deploymentQuota", update.DeploymentQuota)
	addIfNotNil(updateData, "publicKeys", update.PublicKeys)

	if len(updateData) == 0 {
		return nil
	}

	filter := bson.D{{"id", userID}}
	updateDoc := bson.D{{"$set", updateData}}

	_, err := models.UserCollection.UpdateOne(context.Background(), filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update user info for %s. details: %s", userID, err)
	}

	return nil
}

func addIfNotNil(data bson.M, key string, value interface{}) {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return
	}
	data[key] = value
}
