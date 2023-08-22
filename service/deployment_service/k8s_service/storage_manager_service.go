package k8s_service

import (
	"fmt"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
)

func CreateStorageManager(id string, params *storageManagerModel.CreateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create storage manager in k8s. details: %s", err)
	}

	storageManager, err := storageManagerModel.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if storageManager == nil {
		return makeError(fmt.Errorf("storage manager not found, assuming it was deleted"))
	}

	zone := conf.Env.Deployment.GetZone(storageManager.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone not found"))
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	ss := &storageManager.Subsystems.K8s
	updateDb := func(id, subsystem, key string, update interface{}) error {
		return storageManagerModel.UpdateSubsystemByID(id, subsystem, key, update)
	}

	appName := "storage-manager"

	// Namespace
	namespace := &ss.Namespace
	if !ss.Namespace.Created() {
		name := fmt.Sprintf("storage-%s", storageManager.OwnerID)
		namespace, err = createNamespace(client, storageManager.ID, ss, createNamespacePublic(name), updateDb)
		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	if !ss.Deployment.Created() {
		dockerImage := "filebrowser/filebrowser"
		_, err = createK8sDeployment(client, storageManager.ID, ss, createDeploymentPublic(namespace.FullName, appName, dockerImage, nil), updateDb)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	port := 80
	service := &ss.Service
	if !ss.Service.Created() {
		service, err = createService(client, storageManager.ID, ss, createServicePublic(namespace.FullName, appName, &port), updateDb)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	if !ss.Ingress.Created() {
		_, err = createIngress(client, storageManager.ID, ss, createIngressPublic(
			namespace.FullName,
			storageManager.OwnerID,
			service.Name,
			service.Port,
			[]string{getExternalFQDN(storageManager.OwnerID, zone)},
		), updateDb)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}
