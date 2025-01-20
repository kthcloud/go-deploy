package users

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/user_repo"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/utils"
	"github.com/kthcloud/go-deploy/service/v2/users/opts"
)

// Get gets a user
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.User, error) {
	_ = utils.GetFirstOrDefault(opts)

	if c.V2.Auth() != nil && id != c.V2.Auth().User.ID && !c.V2.Auth().User.IsAdmin {
		return nil, nil
	}

	return c.User(id, user_repo.New())
}

// GetByApiKey gets a user by their API key
func (c *Client) GetByApiKey(apiKey string) (*model.User, error) {
	return user_repo.New().WithApiKey(apiKey).Get()
}

// GetUsage gets the usage of a user, such as number of deployments and CPU cores used
func (c *Client) GetUsage(userID string) (*model.UserUsage, error) {
	vmUsage, err := c.V2.VMs().GetUsage(userID)
	if err != nil {
		return nil, err
	}

	deploymentUsage, err := c.V2.Deployments().GetUsage(userID)
	if err != nil {
		return nil, err
	}

	usage := &model.UserUsage{
		CpuCores: float64(vmUsage.CpuCores) + deploymentUsage.CpuCores,
		RAM:      float64(vmUsage.RAM) + deploymentUsage.RAM,
		DiskSize: vmUsage.DiskSize,
	}

	return usage, nil
}

// List lists users
func (c *Client) List(opts ...opts.ListOpts) ([]model.User, error) {
	o := utils.GetFirstOrDefault(opts)

	umc := user_repo.New()

	if o.Pagination != nil {
		umc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.Search != nil {
		umc.WithSearch(*o.Search)
	}

	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin || !o.All {
		user, err := umc.GetByID(c.V2.Auth().User.ID)
		if err != nil {
			return nil, err
		}

		return []model.User{*user}, nil
	}

	return c.Users(umc)
}

// Exists checks if a user exists
//
// This does not use AuthInfo
func (c *Client) Exists(id string) (bool, error) {
	return user_repo.New().ExistsByID(id)
}

// Synchronize creates a user or updates an existing user.
// It does nothing if no auth info is provided.
//
// It does not use AuthInfo as it is meant to be used by the auth service.
func (c *Client) Synchronize(authParams *model.AuthParams) (*model.User, error) {
	var effectiveRole *model.EffectiveRole
	if len(authParams.Roles) > 0 {
		effectiveRole = &model.EffectiveRole{
			Name:        authParams.Roles[len(authParams.Roles)-1].Name,
			Description: authParams.Roles[len(authParams.Roles)-1].Description,
		}
	}

	synchronizeParams := &model.UserSynchronizeParams{
		Username:      authParams.Username,
		FirstName:     authParams.FirstName,
		LastName:      authParams.LastName,
		Email:         authParams.Email,
		IsAdmin:       authParams.IsAdmin,
		EffectiveRole: effectiveRole,
	}

	umc := user_repo.New()

	user, err := umc.Synchronize(authParams.UserID, synchronizeParams)
	if err != nil {
		return nil, err
	}

	if user.Gravatar.FetchedAt.IsZero() || user.Gravatar.FetchedAt.Add(model.FetchGravatarInterval).Before(time.Now()) {
		gravatarURL, err := c.FetchGravatar(user.ID)
		if err != nil {
			return nil, err
		}

		if gravatarURL == nil {
			err = umc.UnsetGravatar(user.ID)
			if err != nil {
				return nil, err
			}
		} else {
			err = umc.SetGravatar(user.ID, *gravatarURL)
			if err != nil {
				return nil, err
			}
		}
	}

	return user, nil
}

// Discover returns a list of users that the requesting user has access to.
//
// It uses search param to enable searching in multiple fields.
// If UserID is provided, it returns a single user.
func (c *Client) Discover(opts ...opts.DiscoverOpts) ([]body.UserReadDiscovery, error) {
	o := utils.GetFirstOrDefault(opts)
	umc := user_repo.New()

	if o.Search != nil {
		umc.WithSearch(*o.Search)
	}

	if o.Pagination != nil {
		umc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	// Single user is requested
	if o.UserID != nil {
		user, err := c.User(*o.UserID, umc)
		if err != nil {
			return nil, err
		}

		if user == nil {
			return nil, nil
		}

		usersRead := []body.UserReadDiscovery{
			{
				ID:          user.ID,
				Username:    user.Username,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Email:       user.Email,
				GravatarURL: user.Gravatar.URL,
			},
		}

		return usersRead, nil
	}

	// Multiple users are requested
	users, err := c.Users(umc)
	if err != nil {
		return nil, err
	}

	var usersRead []body.UserReadDiscovery
	for _, user := range users {
		if c.V2.Auth() != nil && user.ID == c.V2.Auth().User.ID {
			continue
		}

		usersRead = append(usersRead, body.UserReadDiscovery{
			ID:          user.ID,
			Username:    user.Username,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Email:       user.Email,
			GravatarURL: user.Gravatar.URL,
		})
	}

	return usersRead, nil
}

// Update updates a user
func (c *Client) Update(userID string, dtoUserUpdate *body.UserUpdate) (*model.User, error) {
	umc := user_repo.New()

	user, err := c.User(userID, umc)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, sErrors.ErrUserNotFound
	}

	userUpdate := model.UserUpdateParams{}.FromDTO(dtoUserUpdate, user.ApiKeys)

	err = umc.UpdateWithParams(userID, &userUpdate)
	if err != nil {
		return nil, err
	}

	return c.RefreshUser(userID, umc)
}

// FetchGravatar checks if the user has a gravatar image and fetches it if it exists.
// If the user does not have a gravatar image, it returns nil.
func (c *Client) FetchGravatar(userID string) (*string, error) {
	umc := user_repo.New()

	if c.V2.Auth() != nil && userID != c.V2.Auth().User.ID && !c.V2.Auth().User.IsAdmin {
		return nil, nil
	}

	user, err := c.User(userID, umc)
	if err != nil {
		return nil, err
	}

	hasher := sha256.Sum256([]byte(strings.TrimSpace(user.Email)))
	hash := hex.EncodeToString(hasher[:])

	gravatarURL := "https://www.gravatar.com/avatar/" + hash + "?d=404"

	// Check if image exists
	resp, err := http.Head(gravatarURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	// Trim the query string
	gravatarURL = gravatarURL[:strings.Index(gravatarURL, "?")]

	return &gravatarURL, nil
}

func (c *Client) ListTestUsers() ([]model.User, error) {
	return []model.User{
		{
			ID:        model.TestAdminUserID,
			Username:  "tester-admin",
			FirstName: "tester-admin-first",
			LastName:  "tester-admin-last",
			Email:     "tester-admin@test.com",
			IsAdmin:   true,
			EffectiveRole: model.EffectiveRole{
				Name:        getStrongestRole().Name,
				Description: getStrongestRole().Description,
			},
			LastAuthenticatedAt: time.Now(),
			ApiKeys: []model.ApiKey{
				{
					Name:      "test-api-key-admin",
					Key:       model.TestAdminUserApiKey,
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().AddDate(100, 0, 0),
				},
			},
		},
		{
			ID:        model.TestPowerUserID,
			Username:  "tester-power",
			FirstName: "tester-power-first",
			LastName:  "tester-power-last",
			Email:     "tester-power@test.com",
			IsAdmin:   false,
			EffectiveRole: model.EffectiveRole{
				Name:        getStrongestRole().Name,
				Description: getStrongestRole().Description,
			},
			LastAuthenticatedAt: time.Now(),
			ApiKeys: []model.ApiKey{
				{
					Name:      "test-api-key-power",
					Key:       model.TestPowerUserApiKey,
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().AddDate(100, 0, 0),
				},
			},
		},
		{
			ID:        model.TestDefaultUserID,
			Username:  "tester-default",
			FirstName: "tester-default-first",
			LastName:  "tester-default-last",
			Email:     "tester-default@test.com",
			IsAdmin:   false,
			EffectiveRole: model.EffectiveRole{
				Name:        getStrongestRole().Name,
				Description: getStrongestRole().Description,
			},
			LastAuthenticatedAt: time.Now(),
			ApiKeys: []model.ApiKey{
				{
					Name:      "test-api-key-default",
					Key:       model.TestDefaultUserApiKey,
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().AddDate(100, 0, 0),
				},
			},
		},
	}, nil
}

func getStrongestRole() *model.Role {
	return &config.Config.Roles[len(config.Config.Roles)-1]
}
