package k8s_service

import (
	"fmt"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/pkg/conf"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/deployment_service/k8s_service/helpers"
	"go-deploy/utils/subsystemutils"
	"log"
	"path"
)

const (
	storageManagerNamePrefix  = "system"
	storageManagerAppName     = "storage-manager"
	storageManagerAppNameAuth = "storage-manager-auth"
)

func getStorageManagerNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(fmt.Sprintf("%s-%s", storageManagerNamePrefix, userID))
}

func CreateStorageManager(id string, params *storageManagerModel.CreateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create storage manager in k8s. details: %w", err)
	}

	storageManager, err := storageManagerModel.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if storageManager == nil {
		log.Println("storage manager", id, "not found when creating, assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&storageManager.Subsystems.K8s, storageManager.Zone, getStorageManagerNamespaceName(storageManager.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &storageManager.Subsystems.K8s

	initVolumes, volumes := GetStorageManagerVolumes(storageManager.OwnerID, storageManagerAppName)
	allVolumes := append(initVolumes, volumes...)

	jobs := []storageManagerModel.Job{
		{
			Name:    "init",
			Image:   "busybox",
			Command: []string{"/bin/mkdir"},
			Args: []string{
				"-p",
				path.Join("/exports", params.UserID, "data"),
				path.Join("/exports", params.UserID, "user"),
			},
		},
	}

	// Namespace
	if service.NotCreated(&ss.Namespace) {
		public := helpers.CreateNamespacePublic(getStorageManagerNamespaceName(params.UserID))
		_, err = client.CreateNamespace(storageManager.ID, public)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	for _, volume := range allVolumes {
		if service.NotCreated(ss.GetPV(volume.Name)) {
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			nfsPath := path.Join(client.Zone.Storage.NfsParentPath, volume.ServerPath)
			k8sName := fmt.Sprintf("%s-%s", volume.Name, params.UserID)

			public := helpers.CreatePvPublic(k8sName, capacity, nfsPath, client.Zone.Storage.NfsServer)
			_, err = client.CreatePV(storageManager.ID, volume.Name, public)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolumeClaim
	for _, volume := range allVolumes {
		if service.NotCreated(ss.GetPVC(volume.Name)) {
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			pvName := fmt.Sprintf("%s-%s", volume.Name, params.UserID)

			public := helpers.CreatePvcPublic(client.Namespace, volume.Name, capacity, pvName)
			_, err = client.CreatePVC(storageManager.ID, volume.Name, public)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Job
	for _, job := range jobs {
		if service.NotCreated(ss.GetJob(job.Name)) {
			public := helpers.CreateJobPublic(client.Namespace, job.Name, job.Image, job.Command, job.Args, initVolumes)
			_, err = client.CreateJob(storageManager.ID, job.Name, public)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Deployment
	filebrowserAppName := "filebrowser"
	oauthProxyAppName := "oauth-proxy"

	filebrowserPort := 80
	oauthProxyPort := 4180

	if service.NotCreated(ss.GetDeployment(filebrowserAppName)) {
		public := helpers.CreateFileBrowserDeploymentPublic(client.Namespace, storageManagerAppName, volumes, nil)
		_, err = client.CreateK8sDeployment(storageManager.ID, filebrowserAppName, public)
		if err != nil {
			return makeError(err)
		}
	}

	if service.NotCreated(ss.GetDeployment(oauthProxyAppName)) {
		public := helpers.CreateOAuthProxyDeploymentPublic(client.Namespace, storageManagerAppNameAuth, params.UserID, client.Zone)
		_, err = client.CreateK8sDeployment(storageManager.ID, oauthProxyAppName, public)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	if service.NotCreated(ss.GetService(filebrowserAppName)) {
		public := helpers.CreateServicePublic(client.Namespace, storageManagerAppName, filebrowserPort, filebrowserPort)
		_, err = client.CreateService(storageManager.ID, filebrowserAppName, public)
		if err != nil {
			return makeError(err)
		}
	}

	if service.NotCreated(ss.GetService(oauthProxyAppName)) {
		public := helpers.CreateServicePublic(client.Namespace, storageManagerAppNameAuth, oauthProxyPort, oauthProxyPort)
		_, err = client.CreateService(storageManager.ID, oauthProxyAppName, public)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	if service.NotCreated(ss.GetIngress(filebrowserAppName)) {
		public := helpers.CreateIngressPublic(
			client.Namespace,
			storageManagerAppNameAuth,
			ss.ServiceMap[oauthProxyAppName].Name,
			ss.ServiceMap[oauthProxyAppName].Port,
			GetStorageManagerExternalFQDN(storageManager.OwnerID, client.Zone),
		)
		_, err = client.CreateIngress(storageManager.ID, oauthProxyAppName, public)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func DeleteStorageManager(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete storage manager in k8s. details: %w", err)
	}

	log.Println("deleting k8s for storage manager", id)

	storageManager, err := storageManagerModel.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if storageManager == nil {
		log.Println("storage manager", id, "not found when deleting, assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&storageManager.Subsystems.K8s, storageManager.Zone, getStorageManagerNamespaceName(storageManager.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &storageManager.Subsystems.K8s

	// Deployment
	for mapName := range ss.DeploymentMap {
		err = client.DeleteK8sDeployment(storageManager.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for mapName := range ss.ServiceMap {
		err = client.DeleteService(storageManager.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	for mapName := range ss.IngressMap {
		err = client.DeleteIngress(storageManager.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// Job
	for mapName := range ss.JobMap {
		err = client.DeleteJob(storageManager.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for mapName := range ss.PvcMap {
		err = client.DeletePVC(storageManager.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	for mapName := range ss.PvMap {
		err = client.DeletePV(storageManager.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func RepairStorageManager(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair storage manager in k8s. details: %w", err)
	}

	storageManager, err := storageManagerModel.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if storageManager == nil {
		log.Println("storage manager", id, "not found when repairing, assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&storageManager.Subsystems.K8s, storageManager.Zone, getStorageManagerNamespaceName(storageManager.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &storageManager.Subsystems.K8s

	// deployment
	for mapName, _ := range ss.DeploymentMap {
		err = client.RepairK8sDeployment(storageManager.ID, mapName, func() *k8sModels.DeploymentPublic {
			if mapName == "filebrowser" {
				_, volumes := GetStorageManagerVolumes(storageManager.OwnerID, storageManagerAppName)
				return helpers.CreateFileBrowserDeploymentPublic(client.Namespace, mapName, volumes, nil)
			} else if mapName == "oauth-proxy" {
				return helpers.CreateOAuthProxyDeploymentPublic(client.Namespace, mapName, storageManager.OwnerID, client.Zone)
			}
			return nil
		})
		if err != nil {
			return makeError(err)
		}
	}

	// service
	for mapName, _ := range ss.ServiceMap {
		err = client.RepairService(storageManager.ID, mapName, func() *k8sModels.ServicePublic {
			if mapName == "filebrowser" {
				return helpers.CreateServicePublic(client.Namespace, mapName, 80, 80)
			} else if mapName == "oauth-proxy" {
				return helpers.CreateServicePublic(client.Namespace, mapName, 4180, 4180)
			}
			return nil
		})
		if err != nil {
			return makeError(err)
		}
	}

	// ingress
	for mapName, _ := range ss.IngressMap {
		err = client.RepairIngress(storageManager.ID, mapName, func() *k8sModels.IngressPublic {
			if mapName == "oauth-proxy" {
				k8sService := ss.GetService(mapName)
				if service.NotCreated(k8sService) {
					return nil
				}

				return helpers.CreateIngressPublic(
					client.Namespace,
					mapName,
					k8sService.Name,
					k8sService.Port,
					GetStorageManagerExternalFQDN(storageManager.OwnerID, client.Zone),
				)
			}
			return nil
		})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}
