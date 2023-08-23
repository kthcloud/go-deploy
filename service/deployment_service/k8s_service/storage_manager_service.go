package k8s_service

import (
	"fmt"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"log"
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
		log.Println("storage manager", id, "not found when creating, assuming it was deleted")
		return nil
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
		namespace, err = createNamespace(client, storageManager.ID, ss, createNamespacePublic(name), storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	if ss.PvMap == nil {
		ss.PvMap = make(map[string]k8sModels.PvPublic)
	}

	allVolumes := append(volumes, initVolumes...)

	for _, volume := range allVolumes {
		pv, exists := ss.PvMap[volume.Name]
		k8sName := fmt.Sprintf("%s-%s", volume.Name, params.UserID)
		if !pv.Created() || !exists {
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			nfsPath := path.Join(zone.Storage.NfsParentPath, volume.ServerPath)

			public := createPvPublic(k8sName, capacity, nfsPath, zone.Storage.NfsServer)
			_, err = createPV(client, storageManager.ID, volume.Name, ss, public, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolumeClaim
	if ss.PvcMap == nil {
		ss.PvcMap = make(map[string]k8sModels.PvcPublic)
	}

	for _, volume := range allVolumes {
		pvc, exists := ss.PvcMap[volume.Name]
		pvName := fmt.Sprintf("%s-%s", volume.Name, params.UserID)
		if !pvc.Created() || !exists {
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			public := createPvcPublic(namespace.FullName, volume.Name, capacity, pvName)
			_, err = createPVC(client, storageManager.ID, volume.Name, ss, public, storageManagerModel.UpdateSubsystemByID)
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
			_, err = createJob(client, storageManager.ID, j.Name, ss, public, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Deployment
	port := 80
	if !ss.Deployment.Created() {
		public := createStorageManagerDeploymentPublic(namespace.FullName, appName, volumes, nil)
		_, err = createK8sDeployment(client, storageManager.ID, ss, public, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	service := &ss.Service
	if !ss.Service.Created() {
		public := createServicePublic(namespace.FullName, appName, port)
		service, err = createService(client, storageManager.ID, ss, public, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	if !ss.Ingress.Created() {
		public := createIngressPublic(
			namespace.FullName,
			appName,
			service.Name,
			service.Port,
			[]string{getStorageManagerExternalFQDN(storageManager.OwnerID, zone)},
		)
		_, err = createIngress(client, storageManager.ID, ss, public, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func DeleteStorageManager(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete storage manager in k8s. details: %s", err)
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

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	ss := &storageManager.Subsystems.K8s

	// Deployment
	if ss.Deployment.Created() {
		err = deleteDeployment(client, storageManager.ID, ss, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	if ss.Service.Created() {
		err = deleteService(client, storageManager.ID, ss, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	if ss.Ingress.Created() {
		err = deleteIngress(client, storageManager.ID, ss, storageManagerModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Job
	for jobName, job := range ss.JobMap {
		if job.Created() {
			err = deleteJob(client, storageManager.ID, jobName, ss, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolumeClaim
	for pvcName, pvc := range ss.PvcMap {
		if pvc.Created() {
			err = deletePVC(client, storageManager.ID, pvcName, ss, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolume
	for pvName, pv := range ss.PvMap {
		if pv.Created() {
			err = deletePV(client, storageManager.ID, pvName, ss, storageManagerModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	return nil
}

func RepairStorageManager(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair storage manager in k8s. details: %s", err)
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

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	ss := &storageManager.Subsystems.K8s

	// deployment
	err = repairDeployment(client, storageManager.ID, ss, storageManagerModel.UpdateSubsystemByID)
	if err != nil {
		return makeError(err)
	}

	// service
	err = repairService(client, storageManager.ID, ss, storageManagerModel.UpdateSubsystemByID)
	if err != nil {
		return makeError(err)
	}

	// ingres
	err = repairIngress(client, storageManager.ID, ss, storageManagerModel.UpdateSubsystemByID)
	if err != nil {
		return makeError(err)
	}

	return nil
}
