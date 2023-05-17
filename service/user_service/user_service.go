package user_service

import (
	"go-deploy/models/dto/body"
	userModel "go-deploy/models/sys/user"
	"go-deploy/pkg/conf"
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

func GetOrCreate(userID, username string, roles []string) (*userModel.User, error) {
	err := userModel.Create(userID, username, roles)
	if err != nil {
		return nil, err
	}

	return userModel.GetByID(userID)
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

func GetQuotaByUserID(id string) (*userModel.Quota, error) {
	user, err := userModel.GetByID(id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	quota := conf.Env.GetQuota(user.Roles)

	if quota == nil {
		return nil, nil
	}

	return &userModel.Quota{
		Deployments: quota.Deployments,
		CpuCores:    quota.CpuCores,
		RAM:         quota.RAM,
		DiskSize:    quota.DiskSize,
	}, nil
}
