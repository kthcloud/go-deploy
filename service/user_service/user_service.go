package user_service

import (
	"go-deploy/models/dto"
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

func Update(requestedUserID, userID string, isAdmin bool, dtoUserUpdate dto.UserUpdate) error {
	if !isAdmin && requestedUserID != userID {
		return nil
	}

	userUpdate := &userModel.UserUpdate{
		Username:   dtoUserUpdate.Username,
		Email:      dtoUserUpdate.Email,
		PublicKeys: dtoUserUpdate.PublicKeys,
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