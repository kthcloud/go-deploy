package user_service

import (
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	userModel "go-deploy/models/sys/user"
	"go-deploy/service"
	"sort"
)

func GetByIdAuth(userID string, auth *service.AuthInfo) (*userModel.User, error) {
	if userID != auth.UserID && !auth.IsAdmin {
		return nil, nil
	}

	return userModel.New().GetByID(userID)
}

func GetByID(userID string) (*userModel.User, error) {
	return userModel.New().GetByID(userID)
}

func Create(auth *service.AuthInfo) (*userModel.User, error) {
	roleNames := make([]string, len(auth.Roles))
	for i, role := range auth.Roles {
		roleNames[i] = role.Name
	}

	effectiveRole := auth.GetEffectiveRole()

	params := &userModel.CreateParams{
		Username:  auth.GetUsername(),
		FirstName: auth.GetFirstName(),
		LastName:  auth.GetLastName(),
		Email:     auth.GetEmail(),
		IsAdmin:   auth.IsAdmin,
		EffectiveRole: &userModel.EffectiveRole{
			Name:        effectiveRole.Name,
			Description: effectiveRole.Description,
		},
	}

	err := userModel.New().Create(auth.UserID, params)
	if err != nil {
		return nil, err
	}

	return userModel.New().GetByID(auth.UserID)
}

func DiscoverAuth(search *string, auth *service.AuthInfo) ([]body.UserReadDiscovery, error) {
	client := userModel.New()

	if search != nil {
		client.AddSearch(*search)
	}

	users, err := client.ListAll()
	if err != nil {
		return nil, err
	}

	var usersRead []body.UserReadDiscovery
	for _, user := range users {
		if user.ID == auth.UserID {
			continue
		}

		usersRead = append(usersRead, body.UserReadDiscovery{
			ID:        user.ID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		})
	}

	sort.Slice(usersRead, func(i, j int) bool {
		return usersRead[i].FirstName < usersRead[j].FirstName
	})

	return usersRead, nil
}

func ListAuth(allUsers bool, search *string, auth *service.AuthInfo, pagination *query.Pagination) ([]userModel.User, error) {
	client := userModel.New()

	if pagination != nil {
		client.AddPagination(pagination.Page, pagination.PageSize)
	}

	if search != nil {
		client.AddSearch(*search)
	}

	if !allUsers || (allUsers && !auth.IsAdmin) {
		user, err := client.GetByID(auth.UserID)
		if err != nil {
			return nil, err
		}

		return []userModel.User{*user}, nil
	}

	return client.ListAll()
}

func UpdatedAuth(userID string, dtoUserUpdate *body.UserUpdate, auth *service.AuthInfo) (*userModel.User, error) {
	client := userModel.New()

	if userID != auth.UserID && !auth.IsAdmin {
		return nil, nil
	}

	var publicKeys *[]userModel.PublicKey
	if dtoUserUpdate.PublicKeys != nil {
		k := make([]userModel.PublicKey, len(*dtoUserUpdate.PublicKeys))
		for i, key := range *dtoUserUpdate.PublicKeys {
			k[i] = userModel.PublicKey{
				Name: key.Name,
				Key:  key.Key,
			}
		}

		publicKeys = &k
	}

	userUpdate := &userModel.UpdateParams{
		PublicKeys: publicKeys,
		Onboarded:  dtoUserUpdate.Onboarded,
	}

	err := client.UpdateWithParams(userID, userUpdate)
	if err != nil {
		return nil, err
	}

	return client.GetByID(userID)
}
