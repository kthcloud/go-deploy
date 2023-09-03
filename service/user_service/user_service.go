package user_service

import (
	"go-deploy/models/dto/body"
	userModel "go-deploy/models/sys/user"
	"go-deploy/service"
)

func GetByID(userID string, auth *service.AuthInfo) (*userModel.User, error) {
	if userID != auth.UserID && !auth.IsAdmin {
		return nil, nil
	}

	return userModel.New().GetByID(userID)
}

func GetOrCreate(auth *service.AuthInfo) (*userModel.User, error) {
	roleNames := make([]string, len(auth.Roles))
	for i, role := range auth.Roles {
		roleNames[i] = role.Name
	}

	effectiveRole := auth.GetEffectiveRole()

	err := userModel.New().Create(auth.UserID, auth.GetUsername(), auth.GetEmail(), auth.IsAdmin, &userModel.EffectiveRole{
		Name:        effectiveRole.Name,
		Description: effectiveRole.Description,
	})
	if err != nil {
		return nil, err
	}

	return userModel.New().GetByID(auth.UserID)
}

func GetAll(auth *service.AuthInfo) ([]userModel.User, error) {
	if auth.IsAdmin {
		return userModel.New().GetAll()
	}

	self, err := userModel.New().GetByID(auth.UserID)
	if err != nil {
		return nil, err
	}

	if self == nil {
		return nil, nil
	}

	return []userModel.User{*self}, nil
}

func Update(userID string, dtoUserUpdate *body.UserUpdate, auth *service.AuthInfo) error {
	if userID != auth.UserID && !auth.IsAdmin {
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
		PublicKeys: &publicKeys,
	}

	err := userModel.New().Update(userID, userUpdate)
	if err != nil {
		return err
	}

	return nil
}
