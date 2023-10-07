package k8s_service

import (
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/deployment_service/resources"
	"go-deploy/utils/subsystemutils"
)

type Context struct {
	base.Context

	Client    *k8s.Client
	Generator *resources.K8sGenerator
}

func NewContext(deploymentID string) (*Context, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error creating context in deployment helper client. details: %w", err)
	}

	baseContext, err := base.NewBaseContext(deploymentID)
	if err != nil {
		return nil, makeError(err)
	}

	k8sClient, err := withClient(baseContext.Zone, getNamespaceName(baseContext.Deployment.OwnerID))
	if err != nil {
		return nil, makeError(err)
	}

	return &Context{
		BaseContext: *baseContext,
		Client:      k8sClient,
		Generator:   resources.PublicGenerator().WithDeployment(baseContext.Deployment).K8s(k8sClient.Namespace),
	}, nil
}

func (c *Context) WithCreateParams(params *deploymentModels.CreateParams) *Context {
	c.CreateParams = params
	c.Generator.WithCreateParams(params)
	return c
}

func (c *Context) WithUpdateParams(params *deploymentModels.UpdateParams) *Context {
	c.UpdateParams = params
	c.Generator.WithUpdateParams(params)
	return c
}

func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

func withClient(zone *enviroment.DeploymentZone, namespace string) (*k8s.Client, error) {
	client, err := k8s.New(zone.Client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return client, nil
}
