package k8s_service

import (
	"fmt"
	"go-deploy/models/config"
	deploymentModels "go-deploy/models/sys/deployment"
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

func NewContext(deploymentID string, overrideOwnerID ...string) (*DeploymentContext, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error creating context in deployment helper client. details: %w", err)
	}

	baseContext, err := base.NewDeploymentBaseContext(deploymentID)
	if err != nil {
		return nil, makeError(err)
	}

	var namespace string
	if len(overrideOwnerID) > 0 {
		namespace = getNamespaceName(overrideOwnerID[0])
	} else {
		namespace = getNamespaceName(baseContext.Deployment.OwnerID)
	}

	k8sClient, err := withClient(baseContext.Zone, namespace)
	if err != nil {
		return nil, makeError(err)
	}

	return &DeploymentContext{
		DeploymentContext: *baseContext,
		Client:            k8sClient,
		Generator:         baseContext.Generator.K8s(k8sClient),
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

func withClient(zone *config.DeploymentZone, namespace string) (*k8s.Client, error) {
	client, err := k8s.New(zone.Client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return client, nil
}
