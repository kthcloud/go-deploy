package users

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/team_repo"
	"go-deploy/pkg/db/resources/user_repo"
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"sort"
)

// Client is the client for the User service.
type Client struct {
	// V1 is a reference to the parent client.
	V1 clients.V1

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// New creates a new User service client.
func New(v1 clients.V1, cache ...*core.Cache) *Client {
	var c *core.Cache
	if len(cache) > 0 {
		c = cache[0]
	} else {
		c = core.NewCache()
	}

	return &Client{
		V1:    v1,
		Cache: c,
	}
}

// User returns the User with the given ID.
// After a successful fetch, the User will be cached.
func (c *Client) User(id string, umc *user_repo.Client) (*model.User, error) {
	user := c.Cache.GetUser(id)
	if user == nil {
		var err error
		user, err = umc.GetByID(id)
		if err != nil {
			return nil, err
		}

		c.Cache.StoreUser(user)
	}

	return user, nil
}

// Users returns a list of Users.
// After a successful fetch, the Users will be cached.
func (c *Client) Users(umc *user_repo.Client) ([]model.User, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	users, err := umc.List()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		c.Cache.StoreUser(&user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].FirstName < users[j].FirstName || (users[i].FirstName == users[j].FirstName && users[i].LastName < users[j].LastName)
	})

	return users, nil
}

// RefreshUser refreshes the User with the given ID.
// After a successful fetch, the User will be cached.
func (c *Client) RefreshUser(id string, umc *user_repo.Client) (*model.User, error) {
	user, err := umc.GetByID(id)
	if err != nil {
		return nil, err
	}

	c.Cache.StoreUser(user)
	return user, nil
}

// Team returns the Team with the given ID.
// After a successful fetch, the Team will be cached.
func (c *Client) Team(id string, tmc *team_repo.Client) (*model.Team, error) {
	team := c.Cache.GetTeam(id)
	if team == nil {
		var err error
		team, err = tmc.GetByID(id)
		if err != nil {
			return nil, err
		}

		c.Cache.StoreTeam(team)
	}

	return team, nil
}

// Teams returns a list of Teams.
// After a successful fetch, the Teams will be cached.
func (c *Client) Teams(tmc *team_repo.Client) ([]model.Team, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	teams, err := tmc.List()
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		c.Cache.StoreTeam(&team)
	}

	sort.Slice(teams, func(i, j int) bool {
		return teams[i].CreatedAt.After(teams[j].CreatedAt)
	})

	return teams, nil
}

// RefreshTeam refreshes the Team with the given ID.
// After a successful fetch, the Team will be cached.
func (c *Client) RefreshTeam(id string, tmc *team_repo.Client) (*model.Team, error) {
	team, err := tmc.GetByID(id)
	if err != nil {
		return nil, err
	}

	c.Cache.StoreTeam(team)
	return team, nil
}
