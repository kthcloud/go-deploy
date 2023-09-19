package k8s_service

import (
	"fmt"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/utils/subsystemutils"
	"log"
	"path"
)

const (
	storageManagerNamePrefix  = "system"
	storageManagerAppName     = "storage-manager"
	storageManagerAppNameAuth = "storage-manager-auth"
)

func getNamespaceName(userID string) string {
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

	zone := conf.Env.Deployment.GetZone(storageManager.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone not found"))
	}

	client, err := k8s.New(zone.Client, getNamespaceName(storageManager.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &storageManager.Subsystems.K8s

	initVolumes, volumes := getStorageManagerVolumes(storageManager.OwnerID, storageManagerAppName)
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
		public := createNamespacePublic(getNamespaceName(params.UserID))
		_, err = createNamespace(client, storageManager.ID, ss, public, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	for _, volume := range allVolumes {
		if service.NotCreated(ss.GetPV(volume.Name)) {
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			nfsPath := path.Join(zone.Storage.NfsParentPath, volume.ServerPath)
			k8sName := fmt.Sprintf("%s-%s", volume.Name, params.UserID)

			public := createPvPublic(k8sName, capacity, nfsPath, zone.Storage.NfsServer)
			_, err = createPV(client, storageManager.ID, volume.Name, ss, public, storageManagerModel.UpdateSubsystemByID)
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

			public := createPvcPublic(client.Namespace, volume.Name, capacity, pvName)
			_, err = createPVC(client, storageManager.ID, volume.Name, ss, public, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Job
	for _, job := range jobs {
		if service.NotCreated(ss.GetJob(job.Name)) {
			public := createJobPublic(client.Namespace, job.Name, job.Image, job.Command, job.Args, initVolumes)
			_, err = createJob(client, storageManager.ID, job.Name, ss, public, storageManagerModel.UpdateSubsystemByID)
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
		public := createFileBrowserDeploymentPublic(client.Namespace, storageManagerAppName, volumes, nil)
		_, err = createK8sDeployment(client, storageManager.ID, filebrowserAppName, ss, public, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	if service.NotCreated(ss.GetDeployment(oauthProxyAppName)) {
		public := createOAuthProxyDeploymentPublic(client.Namespace, storageManagerAppNameAuth, params.UserID, zone)
		_, err = createK8sDeployment(client, storageManager.ID, oauthProxyAppName, ss, public, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	if service.NotCreated(ss.GetService(filebrowserAppName)) {
		public := createServicePublic(client.Namespace, storageManagerAppName, filebrowserPort, filebrowserPort)
		_, err = createService(client, storageManager.ID, filebrowserAppName, ss, public, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	if service.NotCreated(ss.GetService(oauthProxyAppName)) {
		public := createServicePublic(client.Namespace, storageManagerAppNameAuth, oauthProxyPort, oauthProxyPort)
		_, err = createService(client, storageManager.ID, oauthProxyAppName, ss, public, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	if service.NotCreated(ss.GetIngress(filebrowserAppName)) {
		public := createIngressPublic(
			client.Namespace,
			storageManagerAppNameAuth,
			ss.ServiceMap[oauthProxyAppName].Name,
			ss.ServiceMap[oauthProxyAppName].Port,
			[]string{getStorageManagerExternalFQDN(storageManager.OwnerID, zone)},
		)
		_, err = createIngress(client, storageManager.ID, oauthProxyAppName, ss, public, storageManagerModel.UpdateSubsystemByID)
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

	zone := conf.Env.Deployment.GetZone(storageManager.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone not found"))
	}

	client, err := k8s.New(zone.Client, getNamespaceName(storageManager.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &storageManager.Subsystems.K8s

	// Deployment
	for mapName, deployment := range ss.DeploymentMap {
		if deployment.Created() {
			err = deleteK8sDeployment(client, storageManager.ID, mapName, ss, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Service
	for mapName, k8sService := range ss.ServiceMap {
		if k8sService.Created() {
			err = deleteService(client, storageManager.ID, mapName, ss, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Ingress
	for mapName, ingress := range ss.IngressMap {
		if ingress.Created() {
			err = deleteIngress(client, storageManager.ID, mapName, ss, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Job
	for mapName, job := range ss.JobMap {
		if job.Created() {
			err = deleteJob(client, storageManager.ID, mapName, ss, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolumeClaim
	for mapName, pvc := range ss.PvcMap {
		if pvc.Created() {
			err = deletePVC(client, storageManager.ID, mapName, ss, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolume
	for mapName, pv := range ss.PvMap {
		if pv.Created() {
			err = deletePV(client, storageManager.ID, mapName, ss, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
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

	zone := conf.Env.Deployment.GetZone(storageManager.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone not found"))
	}

	client, err := k8s.New(zone.Client, getNamespaceName(storageManager.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &storageManager.Subsystems.K8s

	// deployment
	for mapName, _ := range ss.DeploymentMap {
		err = repairDeployment(client, storageManager.ID, mapName, ss, storageManagerModel.UpdateSubsystemByID, func() *k8sModels.DeploymentPublic {
			if mapName == "filebrowser" {
				_, volumes := getStorageManagerVolumes(storageManager.OwnerID, storageManagerAppName)
				return createFileBrowserDeploymentPublic(client.Namespace, mapName, volumes, nil)
			} else if mapName == "oauth-proxy" {
				return createOAuthProxyDeploymentPublic(client.Namespace, mapName, storageManager.OwnerID, zone)
			}
			return nil
		})
		if err != nil {
			return makeError(err)
		}
	}

	// service
	for mapName, _ := range ss.ServiceMap {
		err = repairService(client, storageManager.ID, mapName, ss, storageManagerModel.UpdateSubsystemByID, func() *k8sModels.ServicePublic {
			if mapName == "filebrowser" {
				return createServicePublic(client.Namespace, mapName, 80, 80)
			} else if mapName == "oauth-proxy" {
				return createServicePublic(client.Namespace, mapName, 4180, 4180)
			}
			return nil
		})
		if err != nil {
			return makeError(err)
		}
	}

	// ingress
	for mapName, _ := range ss.IngressMap {
		err = repairIngress(client, storageManager.ID, mapName, ss, storageManagerModel.UpdateSubsystemByID, func() *k8sModels.IngressPublic {
			if mapName == "oauth-proxy" {
				k8sService := ss.GetService(mapName)
				if service.NotCreated(k8sService) {
					return nil
				}

				return createIngressPublic(
					client.Namespace,
					mapName,
					k8sService.Name,
					k8sService.Port,
					[]string{getStorageManagerExternalFQDN(storageManager.OwnerID, zone)},
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
