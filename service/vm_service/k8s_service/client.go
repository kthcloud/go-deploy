package k8s_service

import (
	"fmt"
	configModels "go-deploy/models/config"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/client"
	"go-deploy/utils/subsystemutils"
)

// Client is the client for the Harbor service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]
}

// New creates a new Client.
// If context is not nil, it will be used to create a new BaseClient.
// Otherwise, an empty context will be created.
func New(context *client.Context) *Client {
	c := &Client{
		BaseClient: client.NewBaseClient[Client](context),
	}
	c.BaseClient.SetParent(c)
	return c
}

// Get is a helper function returns resources that assists with interacting with the subsystem.
// Essentially just collector the VM, client and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *client.Opts) (*vmModel.VM, *k8s.Client, *resources.K8sGenerator, error) {
	var vm *vmModel.VM
	var err error

	if opts.VM != "" {
		vm, err = c.VM(opts.VM, nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if vm == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	var kc *k8s.Client
	if opts.Client {
		var zone *configModels.DeploymentZone
		var userID *string

		if vm != nil {
			if vm.DeploymentZone != nil {
				zone = config.Config.Deployment.GetZone(*vm.DeploymentZone)
			}

			userID = &vm.OwnerID
		}

		kc, err = c.Client(userID, zone)
		if kc == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	var g *resources.K8sGenerator
	if opts.Generator {
		g = c.Generator(vm, kc)
		if g == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	return vm, kc, g, nil
}

// Client returns the K8s service client.
//
// If userID or zone is nil, it will try to use the values from the options.
// If none of them are set, it will panic.
func (c *Client) Client(userID *string, zone *configModels.DeploymentZone) (*k8s.Client, error) {
	// User ID specified in options takes precedence.
	appliedUserID := ""
	if c.UserID != "" {
		appliedUserID = c.UserID
	} else if userID != nil {
		appliedUserID = *userID
	}

	if appliedUserID == "" {
		panic("user id is empty")
	}

	// Zone specified in options takes precedence.
	var appliedZone *configModels.DeploymentZone
	if c.DeploymentZone != nil {
		appliedZone = c.DeploymentZone
	} else if zone != nil {
		appliedZone = zone
	}

	if appliedZone == nil {
		panic("deployment zone is nil")
	}

	return withClient(appliedZone, getNamespaceName(appliedUserID))
}

// Generator returns a K8s generator.
func (c *Client) Generator(vm *vmModel.VM, client *k8s.Client) *resources.K8sGenerator {
	pg := resources.PublicGenerator()

	if vm == nil {
		panic("vm is nil")
	}

	var vmZone *configModels.VmZone
	var deploymentZone *configModels.DeploymentZone
	var userID string

	// User ID specified in options takes precedence.
	if c.UserID != "" {
		userID = c.UserID
	} else {
		userID = vm.OwnerID
	}

	if userID == "" {
		panic("user id is empty")
	}

	// Zone specified in options takes precedence.
	if c.Zone != nil {
		vmZone = c.Zone
	} else {
		vmZone = config.Config.VM.GetZone(vm.Zone)
	}

	if vmZone == nil {
		panic("vm zone is nil")
	}

	// Deployment zone specified in options takes precedence.
	if c.DeploymentZone != nil {
		deploymentZone = c.DeploymentZone
	} else if vm.DeploymentZone != nil {
		deploymentZone = config.Config.Deployment.GetZone(*vm.DeploymentZone)
	}

	if deploymentZone == nil {
		panic("deployment zone is nil")
	}

	return pg.WithVmZone(c.Zone).WithDeploymentZone(c.DeploymentZone).K8s(client)
}

// getNamespaceName returns the namespace name for the user.
func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

// withClient returns a new K8s service client.
func withClient(zone *configModels.DeploymentZone, namespace string) (*k8s.Client, error) {
	k8sClient, err := k8s.New(zone.Client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return k8sClient, nil
}
