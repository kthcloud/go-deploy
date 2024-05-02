package model

import (
	"go-deploy/dto/v1/body"
	"go-deploy/pkg/log"
)

// ToDTO converts a User to a body.UserRead DTO.
func (user *User) ToDTO(effectiveRole *Role, usage *UserUsage, storageURL *string) body.UserRead {
	publicKeys := make([]body.PublicKey, len(user.PublicKeys))
	for i, key := range user.PublicKeys {
		publicKeys[i] = body.PublicKey{
			Name: key.Name,
			Key:  key.Key,
		}
	}

	if usage == nil {
		usage = &UserUsage{}
	}

	if effectiveRole == nil {
		log.Println("Effective role is nil when creating user read for user", user.Username)
		effectiveRole = &Role{
			Name:        "unknown",
			Description: "unknown",
		}
	}

	userData := make([]body.UserData, len(user.UserData))
	for i, data := range user.UserData {
		userData[i] = body.UserData{
			Key:   data.Key,
			Value: data.Value,
		}
	}

	userRead := body.UserRead{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		GravatarURL: user.Gravatar.URL,

		PublicKeys: publicKeys,
		UserData:   userData,

		Role:  effectiveRole.ToDTO(false),
		Admin: user.IsAdmin,

		Quota: effectiveRole.Quotas.ToDTO(),
		Usage: usage.ToDTO(),

		StorageURL: storageURL,
	}

	return userRead
}

// ToDTO converts a Usage to a body.Usage DTO.
func (usage *UserUsage) ToDTO() body.Usage {
	return body.Usage{
		Deployments: usage.Deployments,
		CpuCores:    usage.CpuCores,
		RAM:         usage.RAM,
		DiskSize:    usage.DiskSize,
		Snapshots:   usage.Snapshots,
	}
}

// FromDTO converts a body.UserUpdate DTO to a UserUpdateParams.
func (params UserUpdateParams) FromDTO(userUpdateDTO *body.UserUpdate) UserUpdateParams {
	var publicKeys *[]PublicKey
	if userUpdateDTO.PublicKeys != nil {
		k := make([]PublicKey, len(*userUpdateDTO.PublicKeys))
		for i, key := range *userUpdateDTO.PublicKeys {
			k[i] = PublicKey{
				Name: key.Name,
				Key:  key.Key,
			}
		}

		publicKeys = &k
	}

	params.PublicKeys = publicKeys

	var userData *[]UserData
	if userUpdateDTO.UserData != nil {
		d := make([]UserData, len(*userUpdateDTO.UserData))
		for i, data := range *userUpdateDTO.UserData {
			d[i] = UserData{
				Key:   data.Key,
				Value: data.Value,
			}
		}

		userData = &d
	}

	params.UserData = userData

	return params
}
