package k8s_service

import (
	"fmt"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"path"
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

	initVolumes := []storageManagerModel.Volume{
		{
			Name:       fmt.Sprintf("%s-%s", appName, "init"),
			Init:       false,
			AppPath:    "/exports",
			ServerPath: "",
		},
	}

	volumes := []storageManagerModel.Volume{
		{
			Name:       fmt.Sprintf("%s-%s", appName, "data"),
			Init:       false,
			AppPath:    "/data",
			ServerPath: path.Join(storageManager.OwnerID, "data"),
		},
		{
			Name:       fmt.Sprintf("%s-%s", appName, "user"),
			Init:       false,
			AppPath:    "/deploy",
			ServerPath: path.Join(storageManager.OwnerID, "user"),
		},
	}

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
	namespace := &ss.Namespace
	if !ss.Namespace.Created() {
		name := fmt.Sprintf("system-%s", storageManager.OwnerID)
		namespace, err = createNamespace(client, storageManager.ID, ss, createNamespacePublic(name), updateDb)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	if ss.PvMap == nil {
		ss.PvMap = make(map[string]k8sModels.PvPublic)
	}

	for _, volume := range volumes {
		pv, exists := ss.PvMap[volume.Name]
		k8sName := fmt.Sprintf("%s-%s", volume.Name, params.UserID)
		if !pv.Created() || !exists {
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			nfsPath := path.Join(zone.Storage.NfsParentPath, volume.ServerPath)

			public := createPvPublic(k8sName, capacity, nfsPath, zone.Storage.NfsServer)
			_, err = createPV(client, storageManager.ID, volume.Name, ss, public, updateDb)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolumeClaim
	if ss.PvcMap == nil {
		ss.PvcMap = make(map[string]k8sModels.PvcPublic)
	}

	for _, volume := range volumes {
		pvc, exists := ss.PvcMap[volume.Name]
		pvName := fmt.Sprintf("%s-%s", volume.Name, params.UserID)
		if !pvc.Created() || !exists {
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			public := createPvcPublic(namespace.FullName, volume.Name, capacity, pvName)
			_, err = createPVC(client, storageManager.ID, volume.Name, ss, public, updateDb)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Job
	if ss.JobMap == nil {
		ss.JobMap = make(map[string]k8sModels.JobPublic)
	}

	for _, j := range jobs {
		job, exists := ss.JobMap[j.Name]
		if !job.Created() || !exists {
			public := createJobPublic(namespace.FullName, j.Name, j.Image, j.Command, j.Args, initVolumes)
			_, err = createJob(client, storageManager.ID, j.Name, ss, public, updateDb)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Deployment
	if !ss.Deployment.Created() {
		public := createStorageManagerDeploymentPublic(namespace.FullName, appName, volumes, nil)
		_, err = createK8sDeployment(client, storageManager.ID, ss, public, updateDb)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	port := 80
	service := &ss.Service
	if !ss.Service.Created() {
		service, err = createService(client, storageManager.ID, ss, createServicePublic(namespace.FullName, appName, port), updateDb)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	if !ss.Ingress.Created() {
		_, err = createIngress(client, storageManager.ID, ss, createIngressPublic(
			namespace.FullName,
			appName,
			service.Name,
			service.Port,
			[]string{getStorageManagerExternalFQDN(storageManager.OwnerID, zone)},
		), updateDb)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}
