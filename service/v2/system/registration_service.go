package system

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/host_repo"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
)

func (c *Client) RegisterNode(params *body.HostRegisterParams) error {
	// Validate token
	if params.Token != config.Config.Discovery.Token {
		return sErrors.BadDiscoveryTokenErr
	}

	if params.DisplayName == "" {
		params.DisplayName = params.Name
	}

	// We don't have any mechanism to enable/disable nodes right now
	params.Enabled = true

	// Check if node is schedulable
	zone := config.Config.GetZone(params.Zone)
	if zone == nil {
		return sErrors.ZoneNotFoundErr
	}

	k8sClient, err := c.V2.Deployments().K8s().Client(zone)
	if err != nil {
		return err
	}

	k8sNode, err := k8sClient.ReadNode(params.Name)
	if err != nil {
		return err
	}

	if k8sNode == nil {
		return sErrors.HostNotFoundErr
	}

	params.Schedulable = k8sNode.Schedulable

	// Register node
	err = host_repo.New().Register(model.NewHostByParams(params))
	if err != nil {
		return err
	}

	return nil
}
