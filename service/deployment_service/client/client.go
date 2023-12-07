package client

import (
	"errors"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	roleModel "go-deploy/models/sys/role"
	"go-deploy/pkg/config"
	"go-deploy/service"
	dErrors "go-deploy/service/deployment_service/errors"
	"go-deploy/utils"
)

type ListOptions struct {
	Pagination      *service.Pagination
	GitHubWebhookID int64
	Shared          bool
}

type GetOptions struct {
	TransferCode  string
	HarborWebhook *body.HarborWebhook
	Shared        bool
}

type QuotaOptions struct {
	Quota  *roleModel.Quotas
	Create *body.DeploymentCreate
	Update *body.DeploymentUpdate
}

type UpdateOwnerParams struct {
	OwnerID string
	Params  *body.DeploymentUpdateOwner
}

type Client[parent any] struct {
	p *parent

	id     string
	IDs    []string
	Name   string
	UserID string

	deployment *deploymentModel.Deployment
	Zone       *configModels.DeploymentZone

	Auth *service.AuthInfo
}

func (client *Client[parent]) SetParent(p *parent) {
	client.p = p
}

func (client *Client[parent]) Deployment() *deploymentModel.Deployment {
	if client.deployment == nil {
		err := client.Fetch()
		if err != nil {
			if errors.Is(err, dErrors.DeploymentNotFoundErr) {
				return nil
			}

			utils.PrettyPrintError(err)
			return nil
		}
	}

	return client.deployment
}

func (client *Client[parent]) ID() string {
	if client.id != "" {
		return client.id
	}

	if client.Deployment() != nil {
		return client.Deployment().ID
	}

	return ""
}

func (client *Client[parent]) HasID() bool {
	return client.ID() != ""
}

func (client *Client[parent]) Fetch() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch deployment in service client: %w", err)
	}

	var deployment *deploymentModel.Deployment
	if client.id != "" {
		var err error
		deployment, err = deploymentModel.New().GetByID(client.id)
		if err != nil {
			return makeError(err)
		}
	} else if client.Name != "" {
		var err error
		deployment, err = deploymentModel.New().GetByName(client.Name)
		if err != nil {
			return makeError(err)
		}
	}

	if deployment == nil {
		return makeError(dErrors.DeploymentNotFoundErr)
	}

	zone := config.Config.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return makeError(dErrors.ZoneNotFoundErr)
	}

	client.Zone = zone
	client.deployment = deployment

	return nil
}

func (client *Client[parent]) WithID(id string) *parent {
	client.id = id
	return client.p
}

func (client *Client[parent]) WithIDs(ids []string) *parent {
	client.IDs = ids
	return client.p
}

func (client *Client[parent]) WithName(name string) *parent {
	client.Name = name
	return client.p
}

func (client *Client[parent]) WithUserID(userID string) *parent {
	client.UserID = userID
	return client.p
}

func (client *Client[parent]) WithZone(zone string) *parent {
	client.Zone = config.Config.Deployment.GetZone(zone)
	return client.p
}

func (client *Client[parent]) WithAuth(auth *service.AuthInfo) *parent {
	client.Auth = auth
	return client.p
}
