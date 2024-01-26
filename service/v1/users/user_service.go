package users

import (
	"go-deploy/models/dto/v1/body"
	userModels "go-deploy/models/sys/user"
	"go-deploy/service/utils"
	"go-deploy/service/v1/users/opts"
)

// Get gets a user
//
// It uses service.AuthInfo to only return the resource the requesting user has access to
func (c *Client) Get(id string, opts ...opts.GetOpts) (*userModels.User, error) {
	_ = utils.GetFirstOrDefault(opts)

	if c.V1.Auth() != nil && id != c.V1.Auth().UserID && !c.V1.Auth().IsAdmin {
		return nil, nil
	}

	return c.User(id, userModels.New())
}

// List lists users
//
// It uses service.AuthInfo to only return the resources the requesting user has access to
// It uses the search param to enable searching in multiple fields
func (c *Client) List(opts ...opts.ListOpts) ([]userModels.User, error) {
	o := utils.GetFirstOrDefault(opts)

	umc := userModels.New()

	if o.Pagination != nil {
		umc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.Search != nil {
		umc.WithSearch(*o.Search)
	}

	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin || !o.All {
		user, err := umc.GetByID(c.V1.Auth().UserID)
		if err != nil {
			return nil, err
		}

		return []userModels.User{*user}, nil
	}

	return c.Users(umc)
}

// Exists checks if a user exists
//
// This does not use AuthInfo
func (c *Client) Exists(id string) (bool, error) {
	return userModels.New().ExistsByID(id)
}

// Create creates a user
//
// It uses service.AuthInfo to get the information about the user
// and acts as a synchronization if the user already exists.
func (c *Client) Create() (*userModels.User, error) {
	if !c.V1.HasAuth() {
		return nil, nil
	}

	roleNames := make([]string, len(c.V1.Auth().Roles))
	for i, role := range c.V1.Auth().Roles {
		roleNames[i] = role.Name
	}

	effectiveRole := c.V1.Auth().GetEffectiveRole()

	params := &userModels.CreateParams{
		Username:  c.V1.Auth().GetUsername(),
		FirstName: c.V1.Auth().GetFirstName(),
		LastName:  c.V1.Auth().GetLastName(),
		Email:     c.V1.Auth().GetEmail(),
		IsAdmin:   c.V1.Auth().IsAdmin,
		EffectiveRole: &userModels.EffectiveRole{
			Name:        effectiveRole.Name,
			Description: effectiveRole.Description,
		},
	}

	return userModels.New().Create(c.V1.Auth().UserID, params)
}

// Discover returns a list of users that the requesting user has access to
//
// It uses search param to enable searching in multiple fields
func (c *Client) Discover(opts ...opts.DiscoverOpts) ([]body.UserReadDiscovery, error) {
	o := utils.GetFirstOrDefault(opts)
	umc := userModels.New()

	if o.Search != nil {
		umc.WithSearch(*o.Search)
	}

	if o.Pagination != nil {
		umc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	users, err := c.Users(umc)
	if err != nil {
		return nil, err
	}

	var usersRead []body.UserReadDiscovery
	for _, user := range users {
		if c.V1.Auth() != nil && user.ID == c.V1.Auth().UserID {
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

	return usersRead, nil
}

// Update updates a user
//
// It uses service.AuthInfo to only update the resource the requesting user has access to
func (c *Client) Update(userID string, dtoUserUpdate *body.UserUpdate) (*userModels.User, error) {
	umc := userModels.New()

	if c.V1.Auth() != nil && userID != c.V1.Auth().UserID && !c.V1.Auth().IsAdmin {
		return nil, nil
	}

	var publicKeys *[]userModels.PublicKey
	if dtoUserUpdate.PublicKeys != nil {
		k := make([]userModels.PublicKey, len(*dtoUserUpdate.PublicKeys))
		for i, key := range *dtoUserUpdate.PublicKeys {
			k[i] = userModels.PublicKey{
				Name: key.Name,
				Key:  key.Key,
			}
		}

		publicKeys = &k
	}

	userUpdate := &userModels.UpdateParams{
		PublicKeys: publicKeys,
		Onboarded:  dtoUserUpdate.Onboarded,
	}

	err := umc.UpdateWithParams(userID, userUpdate)
	if err != nil {
		return nil, err
	}

	return c.RefreshUser(userID, umc)
}
