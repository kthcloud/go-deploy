package api_key

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/user_repo"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/utils"
	"time"
)

// Create generates a new API key for the user.
func (c *Client) Create(userID string, dtoApiKeyCreate *body.ApiKeyCreate) (*model.ApiKey, error) {
	key := c.generateKey()

	user, err := c.V2.Users().Get(userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, sErrors.UserNotFoundErr
	}

	params := model.ApiKeyCreateParams{}.FromDTO(dtoApiKeyCreate, key)

	apiKey := &model.ApiKey{
		Name:      params.Name,
		Key:       params.Key,
		CreatedAt: time.Now(),
		ExpiresAt: params.ExpiresAt,
	}

	// Check duplicate name for the API key
	for _, k := range user.ApiKeys {
		if k.Name == apiKey.Name {
			return nil, sErrors.ApiKeyNameTakenErr
		}
	}

	apiKeys := append(user.ApiKeys, *apiKey)
	err = user_repo.New().UpdateWithParams(userID, &model.UserUpdateParams{ApiKeys: &apiKeys})
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

// List returns all API keys for the user.
func (c *Client) List(userID string) ([]model.ApiKey, error) {
	user, err := c.V2.Users().Get(userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, sErrors.UserNotFoundErr
	}

	return user.ApiKeys, nil
}

// Delete removes an API key from the user.
func (c *Client) Delete(userID string, name string) error {
	user, err := c.V2.Users().Get(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return sErrors.UserNotFoundErr
	}

	apiKeys := make([]model.ApiKey, 0)
	for _, key := range user.ApiKeys {
		if key.Name != name {
			apiKeys = append(apiKeys, key)
		}
	}

	return user_repo.New().UpdateWithParams(userID, &model.UserUpdateParams{ApiKeys: &apiKeys})
}

// generateKey generates a new API key for the user.
func (c *Client) generateKey() string {
	return "go_deploy_" + utils.HashStringAlphanumeric(utils.GenerateSalt())
}
