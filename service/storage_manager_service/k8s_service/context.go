package k8s_service

import (
	"fmt"
	"go-deploy/models/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
)

type StorageManagerContext struct {
	base.StorageManagerContext

	Client    *k8s.Client
	Generator *resources.K8sGenerator
}

func NewStorageManagerContext(storageManagerID string) (*StorageManagerContext, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error creating context in deployment helper client. details: %w", err)
	}

	baseContext, err := base.NewStorageManagerBaseContext(storageManagerID)
	if err != nil {
		return nil, makeError(err)
	}

	k8sClient, err := withClient(baseContext.Zone, getNamespace(baseContext.StorageManager.OwnerID))
	if err != nil {
		return nil, makeError(err)
	}

	return &StorageManagerContext{
		StorageManagerContext: *baseContext,
		Client:                k8sClient,
		Generator:             baseContext.Generator.K8s(k8sClient),
	}, nil
}

func getNamespace(userID string) string {
	return subsystemutils.GetPrefixedName("system-" + userID)
}

func withClient(zone *config.DeploymentZone, namespace string) (*k8s.Client, error) {
	client, err := k8s.New(zone.Client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return client, nil
}
