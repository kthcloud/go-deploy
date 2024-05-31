package system

import (
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/host_repo"
	sErrors "go-deploy/service/errors"
)

func (c *Client) RegisterNode(params *body.HostRegisterParams) error {
	// Validate token
	if params.Token != config.Config.Discovery.Token {
		return sErrors.BadDiscoveryTokenErr
	}

	if params.DisplayName == "" {
		params.DisplayName = params.Name
	}

	// Register node
	err := host_repo.New().Register(model.NewHostByParams(params))
	if err != nil {
		return err
	}

	return nil
}
