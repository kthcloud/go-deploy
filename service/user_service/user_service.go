package user_service

import (
	"go-deploy/models/dto/body"
	userModel "go-deploy/models/sys/user"
	"go-deploy/service"
)

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

func GetOrCreate(auth *service.AuthInfo) (*userModel.User, error) {
	roleNames := make([]string, len(auth.Roles))
	for i, role := range auth.Roles {
		roleNames[i] = role.Name
	}

	effectiveRole := auth.GetEffectiveRole()

	err := userModel.Create(auth.UserID, auth.GetUsername(), auth.GetEmail(), auth.IsAdmin, &userModel.EffectiveRole{
		Name:        effectiveRole.Name,
		Description: effectiveRole.Description,
	})
	if err != nil {
		return nil, err
	}

	return userModel.GetByID(auth.UserID)
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

	err := userModel.Update(requestedUserID, userUpdate)
	if err != nil {
		return err
	}

	return nil
}
