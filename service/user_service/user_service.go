package user_service

import (
	"go-deploy/models/dto/body"
	"go-deploy/models/sys/user"
	"go-deploy/pkg/auth"
)

func CreateUser(userID, username string) error {
	return user.Create(userID, username)
}

func GetByID(requestedUserID, userID string, isAdmin bool) (*user.User, error) {
	user, err := user.GetByID(requestedUserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	if !isAdmin && user.ID != userID {
		return nil, nil
	}

	return user, nil
}

func GetOrCreate(token *auth.KeycloakToken) (*user.User, error) {
	err := user.Create(token.Sub, token.PreferredUsername)
	if err != nil {
		return nil, err
	}

	return user.GetByID(token.Sub)
}

func GetAll() ([]user.User, error) {
	return user.GetAll()
}

func Update(requestedUserID, userID string, isAdmin bool, dtoUserUpdate *body.UserUpdate) error {
	if !isAdmin && requestedUserID != userID {
		return nil
	}

	if dtoUserUpdate.PublicKeys == nil {
		dtoUserUpdate.PublicKeys = &[]body.PublicKey{}
	}

	publicKeys := make([]user.PublicKey, len(*dtoUserUpdate.PublicKeys))
	for i, key := range *dtoUserUpdate.PublicKeys {
		publicKeys[i] = user.PublicKey{
			Name: key.Name,
			Key:  key.Key,
		}
	}

	userUpdate := &user.UserUpdate{
		Username:   dtoUserUpdate.Username,
		Email:      dtoUserUpdate.Email,
		PublicKeys: &publicKeys,
	}

	if isAdmin {
		userUpdate.VmQuota = dtoUserUpdate.VmQuota
		userUpdate.DeploymentQuota = dtoUserUpdate.DeploymentQuota
	}

	err := user.Update(requestedUserID, userUpdate)
	if err != nil {
		return err
	}

	return nil
}
