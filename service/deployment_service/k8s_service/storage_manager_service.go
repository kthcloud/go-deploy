package k8s_service

import (
	"errors"
	"fmt"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/resources"
	"golang.org/x/exp/slices"
	"log"
)

func CreateStorageManager(smID string, params *storageManagerModel.CreateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create storage manager in k8s. details: %w", err)
	}

	context, err := NewStorageManagerContext(smID)
	if err != nil {
		if errors.Is(err, base.StorageManagerDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	context.WithCreateParams(params)

	// Namespace
	err = resources.SsCreator(context.Client.CreateNamespace).
		WithDbFunc(dbFuncSM(smID, "namespace")).
		WithPublic(context.Generator.StorageManagerNamespace()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// PersistentVolume
	for _, pvPublic := range context.Generator.PVs() {
		err = resources.SsCreator(context.Client.CreatePV).
			WithDbFunc(dbFuncSM(smID, "pvMap."+pvPublic.Name)).
			WithPublic(&pvPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for _, pvcPublic := range context.Generator.PVCs() {
		err = resources.SsCreator(context.Client.CreatePVC).
			WithDbFunc(dbFuncSM(smID, "pvcMap."+pvcPublic.Name)).
			WithPublic(&pvcPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Job
	for _, jobPublic := range context.Generator.Jobs() {
		err = resources.SsCreator(context.Client.CreateJob).
			WithDbFunc(dbFuncSM(smID, "jobMap."+jobPublic.Name)).
			WithPublic(&jobPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deployment := range context.Generator.Deployments() {
		err = resources.SsCreator(context.Client.CreateDeployment).
			WithDbFunc(dbFuncSM(smID, "deploymentMap."+deployment.Name)).
			WithPublic(&deployment).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, k8sService := range context.Generator.Services() {
		err = resources.SsCreator(context.Client.CreateService).
			WithDbFunc(dbFuncSM(smID, "serviceMap."+k8sService.Name)).
			WithPublic(&k8sService).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	for _, ingress := range context.Generator.Ingresses() {
		err = resources.SsCreator(context.Client.CreateIngress).
			WithDbFunc(dbFuncSM(smID, "ingressMap."+ingress.Name)).
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
			WithResourceID(k8sDeployment.ID).
			WithDbFunc(dbFuncSM(id, "deploymentMap."+mapName)).
			Exec()
	}

	// Service
	for mapName, k8sService := range context.StorageManager.Subsystems.K8s.ServiceMap {
		err = resources.SsDeleter(context.Client.DeleteService).
			WithResourceID(k8sService.ID).
			WithDbFunc(dbFuncSM(id, "serviceMap."+mapName)).
			Exec()
	}

	// Ingress
	for mapName, ingress := range context.StorageManager.Subsystems.K8s.IngressMap {
		err = resources.SsDeleter(context.Client.DeleteIngress).
			WithResourceID(ingress.ID).
			WithDbFunc(dbFuncSM(id, "ingressMap."+mapName)).
			Exec()
	}

	// Job
	for mapName, job := range context.StorageManager.Subsystems.K8s.JobMap {
		err = resources.SsDeleter(context.Client.DeleteJob).
			WithResourceID(job.ID).
			WithDbFunc(dbFuncSM(id, "jobMap."+mapName)).
			Exec()
	}

	// PersistentVolumeClaim
	for mapName, pvc := range context.StorageManager.Subsystems.K8s.PvcMap {
		err = resources.SsDeleter(context.Client.DeletePVC).
			WithResourceID(pvc.ID).
			WithDbFunc(dbFuncSM(id, "pvcMap."+mapName)).
			Exec()
	}

	// PersistentVolume
	for mapName, pv := range context.StorageManager.Subsystems.K8s.PvMap {
		err = resources.SsDeleter(context.Client.DeletePV).
			WithResourceID(pv.ID).
			WithDbFunc(dbFuncSM(id, "pvMap."+mapName)).
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
				WithResourceID(context.StorageManager.Subsystems.K8s.DeploymentMap[mapName].ID).
				WithDbFunc(dbFuncSM(id, "deploymentMap."+mapName)).
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
		).WithGenPublic(&deployments[idx]).WithDbFunc(dbFuncSM(id, "deploymentMap."+mapName)).Exec()

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
				WithResourceID(context.StorageManager.Subsystems.K8s.ServiceMap[mapName].ID).
				WithDbFunc(dbFuncSM(id, "k8s.serviceMap."+mapName)).
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
		).WithGenPublic(&services[idx]).WithDbFunc(dbFuncSM(id, "k8s.serviceMap."+mapName)).Exec()

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
				WithResourceID(context.StorageManager.Subsystems.K8s.IngressMap[mapName].ID).
				WithDbFunc(dbFuncSM(id, "k8s.ingressMap."+mapName)).
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
		).WithGenPublic(&ingresses[idx]).WithDbFunc(dbFuncSM(id, "k8s.ingressMap."+mapName)).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func dbFuncSM(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		return storageManagerModel.UpdateSubsystemByID(id, "k8s."+key, data)
	}
}
