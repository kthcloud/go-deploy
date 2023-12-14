package user_service

import (
	"go-deploy/models/dto/body"
	userModel "go-deploy/models/sys/user"
	"sort"
)

// Get gets a user
//
// It uses AuthInfo to only return the resource the requesting user has access to
func (c *Client) Get(id string, opts *GetUserOpts) (*userModel.User, error) {
	if c.Auth != nil && id != c.Auth.UserID && !c.Auth.IsAdmin {
		return nil, nil
	}

	return userModel.New().GetByID(id)
}

// List lists users
//
// It uses AuthInfo to only return the resources the requesting user has access to
// It uses the search param to enable searching in multiple fields
func (c *Client) List(opts *ListUsersOpts) ([]userModel.User, error) {
	client := userModel.New()

	if opts.Pagination != nil {
		client.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
	}

	if opts.Search != nil {
		client.WithSearch(*opts.Search)
	}

	if c.Auth != nil && !c.Auth.IsAdmin {
		user, err := client.GetByID(c.Auth.UserID)
		if err != nil {
			return nil, err
		}

		return []userModel.User{*user}, nil
	}

	return client.List()
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
func (c *Client) Discover(opts *DiscoverUsersOpts) ([]body.UserReadDiscovery, error) {
	client := userModel.New()

	if opts.Search != nil {
		client.WithSearch(*opts.Search)
	}

	if opts.Pagination != nil {
		client.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
	}

	users, err := client.List()
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

	sort.Slice(usersRead, func(i, j int) bool {
		return usersRead[i].FirstName < usersRead[j].FirstName
	})

	return usersRead, nil
}

// Update updates a user
//
// It uses AuthInfo to only update the resource the requesting user has access to
func (c *Client) Update(userID string, dtoUserUpdate *body.UserUpdate) (*userModel.User, error) {
	client := userModel.New()

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

	err := client.UpdateWithParams(userID, userUpdate)
	if err != nil {
		return nil, err
	}

	return client.GetByID(userID)
}
