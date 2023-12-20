package user_service

import (
	"go-deploy/models/dto/body"
	userModel "go-deploy/models/sys/user"
	"go-deploy/service"
)

// Get gets a user
//
// It uses service.AuthInfo to only return the resource the requesting user has access to
func (c *Client) Get(id string, opts ...GetUserOpts) (*userModel.User, error) {
	_ = service.GetFirstOrDefault(opts)

	if c.Auth != nil && id != c.Auth.UserID && !c.Auth.IsAdmin {
		return nil, nil
	}

	return c.User(id, userModel.New())
}

// List lists users
//
// It uses service.AuthInfo to only return the resources the requesting user has access to
// It uses the search param to enable searching in multiple fields
func (c *Client) List(opts ...ListUsersOpts) ([]userModel.User, error) {
	o := service.GetFirstOrDefault(opts)

	umc := userModel.New()

	if o.Pagination != nil {
		umc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.Search != nil {
		umc.WithSearch(*o.Search)
	}

	if c.Auth != nil && !c.Auth.IsAdmin || !o.All {
		user, err := umc.GetByID(c.Auth.UserID)
		if err != nil {
			return nil, err
		}

		return []userModel.User{*user}, nil
	}

	return c.Users(umc)
}

// Exists checks if a user exists
//
// This does not use AuthInfo
func (c *Client) Exists(id string) (bool, error) {
	return userModel.New().ExistsByID(id)
}

func (c *Client) Create() (*userModel.User, error) {
	if c.Auth == nil {
		return nil, nil
	}

	roleNames := make([]string, len(c.Auth.Roles))
	for i, role := range c.Auth.Roles {
		roleNames[i] = role.Name
	}

	effectiveRole := c.Auth.GetEffectiveRole()

	params := &userModel.CreateParams{
		Username:  c.Auth.GetUsername(),
		FirstName: c.Auth.GetFirstName(),
		LastName:  c.Auth.GetLastName(),
		Email:     c.Auth.GetEmail(),
		IsAdmin:   c.Auth.IsAdmin,
		EffectiveRole: &userModel.EffectiveRole{
			Name:        effectiveRole.Name,
			Description: effectiveRole.Description,
		},
	}

	return userModel.New().Create(c.Auth.UserID, params)
}

// Discover returns a list of users that the requesting user has access to
//
// It uses search param to enable searching in multiple fields
func (c *Client) Discover(opts ...DiscoverUsersOpts) ([]body.UserReadDiscovery, error) {
	o := service.GetFirstOrDefault(opts)
	umc := userModel.New()

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
		if c.Auth != nil && user.ID == c.Auth.UserID {
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
func (c *Client) Update(userID string, dtoUserUpdate *body.UserUpdate) (*userModel.User, error) {
	umc := userModel.New()

	if c.Auth != nil && userID != c.Auth.UserID && !c.Auth.IsAdmin {
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

	err := umc.UpdateWithParams(userID, userUpdate)
	if err != nil {
		return nil, err
	}

	return c.RefreshUser(userID, umc)
}
