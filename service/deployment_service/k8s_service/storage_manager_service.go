package k8s_service

import (
	"errors"
	"fmt"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/deployment_service/resources"
	"golang.org/x/exp/slices"
	"log"
)

func CreateStorageManager(id string, params *storageManagerModel.CreateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create storage manager in k8s. details: %w", err)
	}

	context, err := NewStorageManagerContext(id)
	if err != nil {
		if errors.Is(err, base.StorageManagerDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	context.WithCreateParams(params)

	// Namespace
	err = resources.SsCreator(context.Client.CreateNamespace).
		WithID(context.StorageManager.ID).
		WithDbFunc(storageManagerModel.UpdateSubsystemByID).
		WithDbKey("k8s.namespace").
		WithPublic(context.Generator.StorageManagerNamespace()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// PersistentVolume
	for _, pvPublic := range context.Generator.PVs() {
		err = resources.SsCreator(context.Client.CreatePV).
			WithID(context.StorageManager.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.pvMap." + pvPublic.Name).
			WithPublic(&pvPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for _, pvcPublic := range context.Generator.PVCs() {
		err = resources.SsCreator(context.Client.CreatePVC).
			WithID(context.StorageManager.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.pvcMap." + pvcPublic.Name).
			WithPublic(&pvcPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Job
	for _, jobPublic := range context.Generator.Jobs() {
		err = resources.SsCreator(context.Client.CreateJob).
			WithID(context.StorageManager.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.jobMap." + jobPublic.Name).
			WithPublic(&jobPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deployment := range context.Generator.Deployments() {
		err = resources.SsCreator(context.Client.CreateDeployment).
			WithID(context.StorageManager.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.deploymentMap." + deployment.Name).
			WithPublic(&deployment).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, k8sService := range context.Generator.Services() {
		err = resources.SsCreator(context.Client.CreateService).
			WithID(context.StorageManager.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.serviceMap." + k8sService.Name).
			WithPublic(&k8sService).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	for _, ingress := range context.Generator.Ingresses() {
		err = resources.SsCreator(context.Client.CreateIngress).
			WithID(context.StorageManager.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.ingressMap." + ingress.Name).
			WithPublic(&ingress).
			Exec()

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

	context, err := NewStorageManagerContext(id)
	if err != nil {
		if errors.Is(err, base.StorageManagerDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	// Deployment
	for mapName, k8sDeployment := range context.StorageManager.Subsystems.K8s.DeploymentMap {
		err = resources.SsDeleter(context.Client.DeleteDeployment).
			WithID(context.StorageManager.ID).
			WithResourceID(k8sDeployment.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.deploymentMap." + mapName).
			Exec()
	}

	// Service
	for mapName, k8sService := range context.StorageManager.Subsystems.K8s.ServiceMap {
		err = resources.SsDeleter(context.Client.DeleteService).
			WithID(context.StorageManager.ID).
			WithResourceID(k8sService.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.serviceMap." + mapName).
			Exec()
	}

	// Ingress
	for mapName, ingress := range context.StorageManager.Subsystems.K8s.IngressMap {
		err = resources.SsDeleter(context.Client.DeleteIngress).
			WithID(context.StorageManager.ID).
			WithResourceID(ingress.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.ingressMap." + mapName).
			Exec()
	}

	// Job
	for mapName, job := range context.StorageManager.Subsystems.K8s.JobMap {
		err = resources.SsDeleter(context.Client.DeleteJob).
			WithID(context.StorageManager.ID).
			WithResourceID(job.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.jobMap." + mapName).
			Exec()
	}

	// PersistentVolumeClaim
	for mapName, pvc := range context.StorageManager.Subsystems.K8s.PvcMap {
		err = resources.SsDeleter(context.Client.DeletePVC).
			WithID(context.StorageManager.ID).
			WithResourceID(pvc.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.pvcMap." + mapName).
			Exec()
	}

	// PersistentVolume
	for mapName, pv := range context.StorageManager.Subsystems.K8s.PvMap {
		err = resources.SsDeleter(context.Client.DeletePV).
			WithID(context.StorageManager.ID).
			WithResourceID(pv.ID).
			WithDbFunc(storageManagerModel.UpdateSubsystemByID).
			WithDbKey("k8s.pvMap." + mapName).
			Exec()
	}

	return nil
}

func RepairStorageManager(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair storage manager in k8s. details: %w", err)
	}

	context, err := NewStorageManagerContext(id)
	if err != nil {
		if errors.Is(err, base.StorageManagerDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	// deployment
	for mapName := range context.StorageManager.Subsystems.K8s.DeploymentMap {
		deployments := context.Generator.Deployments()
		idx := slices.IndexFunc(deployments, func(d k8sModels.DeploymentPublic) bool { return d.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteDeployment).
				WithID(id).
				WithDbKey("k8s.deploymentMap." + mapName).
				WithDbFunc(storageManagerModel.UpdateSubsystemByID).
				WithResourceID(context.StorageManager.Subsystems.K8s.DeploymentMap[mapName].ID).
				Exec()

			if err != nil {
				return makeError(err)
			}

			continue
		}

		err = resources.SsRepairer(
			context.Client.ReadDeployment,
			context.Client.CreateDeployment,
			context.Client.UpdateDeployment,
			context.Client.DeleteDeployment,
		).WithID(id).WithGenPublic(&deployments[idx]).WithDbKey("k8s.deploymentMap." + mapName).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// service
	for mapName := range context.StorageManager.Subsystems.K8s.ServiceMap {
		services := context.Generator.Services()
		idx := slices.IndexFunc(services, func(s k8sModels.ServicePublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteService).
				WithID(id).
				WithDbKey("k8s.serviceMap." + mapName).
				WithDbFunc(storageManagerModel.UpdateSubsystemByID).
				WithResourceID(context.StorageManager.Subsystems.K8s.ServiceMap[mapName].ID).
				Exec()

			if err != nil {
				return makeError(err)
			}

			continue
		}

		err = resources.SsRepairer(
			context.Client.ReadService,
			context.Client.CreateService,
			context.Client.UpdateService,
			context.Client.DeleteService,
		).WithID(id).WithGenPublic(&services[idx]).WithDbKey("k8s.serviceMap." + mapName).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// ingress
	for mapName := range context.StorageManager.Subsystems.K8s.IngressMap {
		ingresses := context.Generator.Ingresses()
		idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool { return i.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteIngress).
				WithID(id).
				WithDbKey("k8s.ingressMap." + mapName).
				WithDbFunc(storageManagerModel.UpdateSubsystemByID).
				WithResourceID(context.StorageManager.Subsystems.K8s.IngressMap[mapName].ID).
				Exec()

			if err != nil {
				return makeError(err)
			}

			continue
		}

		err = resources.SsRepairer(
			context.Client.ReadIngress,
			context.Client.CreateIngress,
			context.Client.UpdateIngress,
			context.Client.DeleteIngress,
		).WithID(id).WithGenPublic(&ingresses[idx]).WithDbKey("k8s.ingressMap." + mapName).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}
