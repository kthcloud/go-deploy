package k8s_service

import (
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/base"
	"go-deploy/utils/subsystemutils"
)

type Context struct {
	base.VmContext

	DeploymentZone *configModels.DeploymentZone
	Client         *k8s.Client
	Generator      *resources.K8sGenerator
}

func NewContext(vmID string, ownerID ...string) (*Context, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error creating k8s service context. details: %w", err)
	}

	baseContext, err := base.NewVmBaseContext(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	deploymentZone := config.Config.Deployment.GetZone(baseContext.VM.Zone)
	if deploymentZone == nil {
		return nil, makeError(base.DeploymentZoneNotFoundErr)
	}

	var namespace string
	if len(ownerID) > 0 {
		namespace = getNamespaceName(ownerID[0])
	} else {
		namespace = getNamespaceName(baseContext.VM.OwnerID)
	}

	k8sClient, err := withClient(deploymentZone, namespace)
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

func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(fmt.Sprintf("vm-%s", userID))
}

func withClient(zone *configModels.DeploymentZone, namespace string) (*k8s.Client, error) {
	client, err := k8s.New(zone.Client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return client, nil
}
