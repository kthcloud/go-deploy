package user_info

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/conf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserInfo struct {
	Sub             string `json:"sub" bson:"sub"`
	CachedUsername  string `json:"cachedUsername" bson:"cachedUsername"`
	VmQuota         int    `json:"vmQuota" bson:"vmQuota"`
	DeploymentQuota int    `json:"deploymentQuota" bson:"deploymentQuota"`
}

func createDefaultUserInfo(token *auth.KeycloakToken) UserInfo {
	return UserInfo{
		Sub:             token.Sub,
		CachedUsername:  token.PreferredUsername,
		VmQuota:         conf.Env.VM.DefaultQuota,
		DeploymentQuota: conf.Env.App.DefaultQuota,
	}
}

func CreateEmpty(token *auth.KeycloakToken) error {
	currentUserInfo, err := GetBySub(token.Sub)
	if err != nil {
		return err
	}

	if currentUserInfo != nil {
		return nil
	}

	_, err = models.UserInfoCollection.InsertOne(context.TODO(), createDefaultUserInfo(token))
	if err != nil {
		err = fmt.Errorf("failed to create user info for %s. details: %s", token.PreferredUsername, err)
		return err
	}

	return nil
}

func GetBySub(sub string) (*UserInfo, error) {
	var userInfo UserInfo
	filter := bson.D{{"sub", sub}}
	err := models.UserInfoCollection.FindOne(context.TODO(), filter).Decode(&userInfo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch user info by sub %s. details: %s", sub, err)
		return nil, err
	}

	return &userInfo, err
}
