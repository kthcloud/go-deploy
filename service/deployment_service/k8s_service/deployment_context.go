package k8s_service

import (
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
)

type DeploymentContext struct {
	base.DeploymentContext

	Client    *k8s.Client
	Generator *resources.K8sGenerator
}

func NewContext(deploymentID string) (*DeploymentContext, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error creating context in deployment helper client. details: %w", err)
	}

	baseContext, err := base.NewDeploymentBaseContext(deploymentID)
	if err != nil {
		return nil, makeError(err)
	}

	k8sClient, err := withClient(baseContext.Zone, getNamespaceName(baseContext.Deployment.OwnerID))
	if err != nil {
		return nil, makeError(err)
	}

	return &DeploymentContext{
		DeploymentContext: *baseContext,
		Client:            k8sClient,
		Generator:         baseContext.Generator.K8s(k8sClient.Namespace),
	}, nil
}

func (c *DeploymentContext) WithCreateParams(params *deploymentModels.CreateParams) *DeploymentContext {
	c.CreateParams = params
	c.Generator.WithDeploymentCreateParams(params)
	return c
}

func (c *DeploymentContext) WithUpdateParams(params *deploymentModels.UpdateParams) *DeploymentContext {
	c.UpdateParams = params
	c.Generator.WithDeploymentUpdateParams(params)
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
