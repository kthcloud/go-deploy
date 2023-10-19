package k8s_service

import (
	"fmt"
	"go-deploy/models/sys/enviroment"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/base"
	"go-deploy/utils/subsystemutils"
)

type Context struct {
	base.VmContext

	DeploymentZone *enviroment.DeploymentZone
	Client         *k8s.Client
	Generator      *resources.K8sGenerator
}

func NewContext(vmID string) (*Context, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error creating k8s service context. details: %w", err)
	}

	baseContext, err := base.NewVmBaseContext(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	deploymentZone := conf.Env.Deployment.GetZone(baseContext.VM.Zone)
	if deploymentZone == nil {
		return nil, makeError(base.DeploymentZoneNotFoundErr)
	}

	k8sClient, err := withClient(deploymentZone, getNamespaceName(baseContext.VM.OwnerID))
	if err != nil {
		return nil, makeError(err)
	}

	return &Context{
		VmContext:      *baseContext,
		DeploymentZone: deploymentZone,
		Client:         k8sClient,
		Generator:      baseContext.Generator.WithDeploymentZone(deploymentZone).K8s(k8sClient),
	}, nil
}

func (c *Context) WithCreateParams(params *vmModels.CreateParams) *Context {
	c.CreateParams = params
	if c.Generator != nil {
		c.Generator.WithVmCreateParams(params)
	}
	return c
}

func (c *Context) WithUpdateParams(params *vmModels.UpdateParams) *Context {
	c.UpdateParams = params
	if c.Generator != nil {
		c.Generator.WithVmUpdateParams(params)
	}
	return c
}

func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(fmt.Sprintf("vm-%s", userID))
}

func withClient(zone *enviroment.DeploymentZone, namespace string) (*k8s.Client, error) {
	client, err := k8s.New(zone.Client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return client, nil
}
