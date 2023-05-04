package user_service

import (
	"go-deploy/models/dto/body"
	userModel "go-deploy/models/user"
	"go-deploy/pkg/auth"
)

func CreateUser(userID, username string) error {
	return userModel.Create(userID, username)
}

func GetByID(requestedUserID, userID string, isAdmin bool) (*userModel.User, error) {
	user, err := userModel.GetByID(requestedUserID)
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

func GetOrCreate(token *auth.KeycloakToken) (*userModel.User, error) {
	err := userModel.Create(token.Sub, token.PreferredUsername)
	if err != nil {
		return nil, err
	}

	return userModel.GetByID(token.Sub)
}

func GetAll() ([]userModel.User, error) {
	return userModel.GetAll()
}

func Update(requestedUserID, userID string, isAdmin bool, dtoUserUpdate *body.UserUpdate) error {
	if !isAdmin && requestedUserID != userID {
		return nil
	}

	if dtoUserUpdate.PublicKeys == nil {
		dtoUserUpdate.PublicKeys = &[]body.PublicKey{}
	}

	publicKeys := make([]userModel.PublicKey, len(*dtoUserUpdate.PublicKeys))
	for i, key := range *dtoUserUpdate.PublicKeys {
		publicKeys[i] = userModel.PublicKey{
			Name: key.Name,
			Key:  key.Key,
		}
	}

	userUpdate := &userModel.UserUpdate{
		Username:   dtoUserUpdate.Username,
		Email:      dtoUserUpdate.Email,
		PublicKeys: &publicKeys,
	}

	if isAdmin {
		userUpdate.VmQuota = dtoUserUpdate.VmQuota
		userUpdate.DeploymentQuota = dtoUserUpdate.DeploymentQuota
	}

	err := userModel.Update(requestedUserID, userUpdate)
	if err != nil {
		return err
	}

	return nil
}
