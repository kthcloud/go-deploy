package k8s_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/deployment_service/constants"
	"go-deploy/service/resources"
	"golang.org/x/exp/slices"
	"log"
	"strings"
)

func Create(deploymentID string, params *deploymentModel.CreateParams) error {
	log.Println("setting up k8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %w", params.Name, err)
	}

	context, err := NewContext(deploymentID)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	context.WithCreateParams(params)

	// Namespace
	err = resources.SsCreator(context.Client.CreateNamespace).
		WithDbFunc(dbFunc(deploymentID, "namespace")).
		WithPublic(context.Generator.MainNamespace()).
		Exec()
	if err != nil {
		return makeError(err)
	}

	// PersistentVolume
	for _, pvPublic := range context.Generator.PVs() {
		err = resources.SsCreator(context.Client.CreatePV).
			WithDbFunc(dbFunc(deploymentID, "pvMap."+pvPublic.Name)).
			WithPublic(&pvPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for _, pvcPublic := range context.Generator.PVCs() {
		err = resources.SsCreator(context.Client.CreatePVC).
			WithDbFunc(dbFunc(deploymentID, "pvcMap."+pvcPublic.Name)).
			WithPublic(&pvcPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for _, secretPublic := range context.Generator.Secrets() {
		err = resources.SsCreator(context.Client.CreateSecret).
			WithDbFunc(dbFunc(deploymentID, "secretMap."+secretPublic.Name)).
			WithPublic(&secretPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deploymentPublic := range context.Generator.Deployments() {
		err = resources.SsCreator(context.Client.CreateDeployment).
			WithDbFunc(dbFunc(deploymentID, "deploymentMap."+deploymentPublic.Name)).
			WithPublic(&deploymentPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, servicePublic := range context.Generator.Services() {
		err = resources.SsCreator(context.Client.CreateService).
			WithDbFunc(dbFunc(deploymentID, "serviceMap."+servicePublic.Name)).
			WithPublic(&servicePublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress main
	for _, ingressPublic := range context.Generator.Ingresses() {
		err = resources.SsCreator(context.Client.CreateIngress).
			WithDbFunc(dbFunc(deploymentID, "ingressMap."+ingressPublic.Name)).
			WithPublic(&ingressPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func Delete(id string) error {
	log.Println("deleting k8s for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %w", id, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	// Ingress
	for mapName, ingress := range context.Deployment.Subsystems.K8s.IngressMap {
		err = resources.SsDeleter(context.Client.DeleteIngress).
			WithResourceID(ingress.ID).
			WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
			Exec()
	}

	// Service
	for mapName, k8sService := range context.Deployment.Subsystems.K8s.ServiceMap {
		err = resources.SsDeleter(context.Client.DeleteService).
			WithResourceID(k8sService.ID).
			WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
			Exec()
	}

	// Deployment
	for mapName, k8sDeployment := range context.Deployment.Subsystems.K8s.DeploymentMap {
		err = resources.SsDeleter(context.Client.DeleteDeployment).
			WithResourceID(k8sDeployment.ID).
			WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
			Exec()
	}

	// PersistentVolumeClaim
	for mapName, pvc := range context.Deployment.Subsystems.K8s.PvcMap {
		err = resources.SsDeleter(context.Client.DeletePVC).
			WithResourceID(pvc.ID).
			WithDbFunc(dbFunc(id, "pvcMap."+mapName)).
			Exec()
	}

	// PersistentVolume
	for mapName, pv := range context.Deployment.Subsystems.K8s.PvMap {
		err = resources.SsDeleter(context.Client.DeletePV).
			WithResourceID(pv.ID).
			WithDbFunc(dbFunc(id, "pvMap."+mapName)).
			Exec()
	}

	// Job
	for mapName, job := range context.Deployment.Subsystems.K8s.JobMap {
		err = resources.SsDeleter(context.Client.DeleteJob).
			WithResourceID(job.ID).
			WithDbFunc(dbFunc(id, "jobMap."+mapName)).
			Exec()
	}

	// Secret
	for mapName, secret := range context.Deployment.Subsystems.K8s.SecretMap {
		err = resources.SsDeleter(context.Client.DeleteSecret).
			WithResourceID(secret.ID).
			WithDbFunc(dbFunc(id, "secretMap."+mapName)).
			Exec()
	}

	// Namespace
	err = resources.SsDeleter(func(string) error { return nil }).
		WithResourceID(context.Deployment.Subsystems.K8s.Namespace.ID).
		WithDbFunc(dbFunc(id, "namespace")).
		Exec()

	return nil
}

func Update(id string, params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s for deployment %s. details: %w", id, err)
	}

	if *params == (deploymentModel.UpdateParams{}) {
		return nil
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	context.WithUpdateParams(params)

	if params.InternalPort != nil {
		err = updateInternalPort(context)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Envs != nil {
		err = updateEnvs(context)
		if err != nil {
			return makeError(err)
		}
	}

	if params.CustomDomain != nil && (params.Private == nil || !*params.Private) {
		err = updateCustomDomain(context)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Private != nil {
		err = updatePrivate(context)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Volumes != nil {
		err = updateVolumes(context)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Image != nil {
		err = updateImage(context)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func Restart(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s %s. details: %w", name, err)
	}

	context, err := NewContext(name)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	err = context.Client.RestartDeployment(context.Deployment.ID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair k8s %s. details: %w", id, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	namespace := context.Generator.MainNamespace()
	err = resources.SsRepairer(
		context.Client.ReadNamespace,
		context.Client.CreateNamespace,
		context.Client.UpdateNamespace,
		func(string) error { return nil },
	).WithResourceID(namespace.ID).WithDbFunc(dbFunc(id, "namespace")).WithGenPublic(namespace).Exec()

	if err != nil {
		return makeError(err)
	}

	for mapName, k8sDeployment := range context.Deployment.Subsystems.K8s.DeploymentMap {
		deployments := context.Generator.Deployments()
		idx := slices.IndexFunc(deployments, func(d k8sModels.DeploymentPublic) bool { return d.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteDeployment).
				WithResourceID(k8sDeployment.ID).
				WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
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
		).WithResourceID(deployments[idx].ID).WithDbFunc(dbFunc(id, "deploymentMap."+deployments[idx].Name)).WithGenPublic(&deployments[idx]).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	for mapName, k8sService := range context.Deployment.Subsystems.K8s.ServiceMap {
		services := context.Generator.Services()
		idx := slices.IndexFunc(services, func(s k8sModels.ServicePublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteService).
				WithResourceID(k8sService.ID).
				WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
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
		).WithResourceID(services[idx].ID).WithDbFunc(dbFunc(id, "serviceMap."+services[idx].Name)).WithGenPublic(&services[idx]).Exec()
	}

	for mapName, ingress := range context.Deployment.Subsystems.K8s.IngressMap {
		ingresses := context.Generator.Ingresses()
		idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool { return i.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteIngress).
				WithResourceID(ingress.ID).
				WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
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
		).WithResourceID(ingresses[idx].ID).WithDbFunc(dbFunc(id, "ingressMap."+ingresses[idx].Name)).WithGenPublic(&ingresses[idx]).Exec()
	}

	for mapName := range context.Deployment.Subsystems.K8s.SecretMap {
		secrets := context.Generator.Secrets()
		idx := slices.IndexFunc(secrets, func(s k8sModels.SecretPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteSecret).
				WithResourceID(context.Deployment.Subsystems.K8s.SecretMap[mapName].ID).
				WithDbFunc(dbFunc(id, "secretMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}

			continue
		}

		err = resources.SsRepairer(
			context.Client.ReadSecret,
			context.Client.CreateSecret,
			context.Client.UpdateSecret,
			context.Client.DeleteSecret,
		).WithResourceID(secrets[idx].ID).WithDbFunc(dbFunc(id, "secretMap."+secrets[idx].Name)).WithGenPublic(&secrets[idx]).Exec()
	}

	return nil
}

func updateInternalPort(context *DeploymentContext) error {
	services := context.Generator.Services()
	idx := slices.IndexFunc(services, func(i k8sModels.ServicePublic) bool {
		return i.Name == context.Deployment.Name
	})
	if idx == -1 {
		log.Println("main k8s service for deployment", context.Deployment.ID, "not found when updating internal port. assuming it was deleted")
		return nil
	}

	if service.NotCreated(&services[idx]) {
		log.Println("main k8s service for deployment", context.Deployment.ID, "not created yet when updating internal port. assuming it was deleted")
		return nil
	}

	err := resources.SsUpdater(context.Client.UpdateService).
		WithDbFunc(dbFunc(context.Deployment.ID, "serviceMap."+context.Deployment.Name)).
		WithPublic(&services[idx]).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

func updateEnvs(context *DeploymentContext) error {
	deployments := context.Generator.Deployments()
	idx := slices.IndexFunc(deployments, func(i k8sModels.DeploymentPublic) bool {
		return i.Name == context.Deployment.Name
	})
	if idx == -1 {
		log.Println("main k8s deployment for deployment", context.Deployment.ID, "not found when updating envs. assuming it was deleted")
		return nil
	}

	if service.NotCreated(&deployments[idx]) {
		log.Println("main k8s deployment for deployment", context.Deployment.ID, "not created yet when updating envs. assuming it was deleted")
		return nil
	}

	err := resources.SsUpdater(context.Client.UpdateDeployment).
		WithDbFunc(dbFunc(context.Deployment.ID, "deploymentMap."+context.Deployment.Name)).
		WithPublic(&deployments[idx]).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

func updateCustomDomain(context *DeploymentContext) error {
	ingresses := context.Generator.Ingresses()
	idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool {
		return i.Name == constants.CustomDomainSuffix(context.Deployment.Name)
	})
	if idx == -1 {
		return nil
	}

	var err error

	if service.NotCreated(&ingresses[idx]) {
		err = resources.SsCreator(context.Client.CreateIngress).
			WithDbFunc(dbFunc(context.Deployment.ID, "ingressMap."+constants.CustomDomainSuffix(context.Deployment.Name))).
			WithPublic(&ingresses[idx]).
			Exec()
	} else {
		err = resources.SsUpdater(context.Client.UpdateIngress).
			WithDbFunc(dbFunc(context.Deployment.ID, "ingressMap."+constants.CustomDomainSuffix(context.Deployment.Name))).
			WithPublic(&ingresses[idx]).
			Exec()
	}

	if err != nil {
		if strings.Contains(err.Error(), "is already defined in ingress") {
			return base.CustomDomainInUseErr
		}

		return err
	}

	return nil
}

func updatePrivate(context *DeploymentContext) error {
	if context.MainApp.Private {
		for mapName, ingress := range context.Deployment.Subsystems.K8s.IngressMap {
			err := resources.SsDeleter(context.Client.DeleteIngress).
				WithResourceID(ingress.ID).
				WithDbFunc(dbFunc(context.Deployment.ID, "ingressMap."+mapName)).
				Exec()

			if err != nil {
				return err
			}
		}
	} else {
		for _, ingressPublic := range context.Generator.Ingresses() {
			err := resources.SsCreator(context.Client.CreateIngress).
				WithDbFunc(dbFunc(context.Deployment.ID, "ingressMap."+ingressPublic.Name)).
				WithPublic(&ingressPublic).
				Exec()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func updateVolumes(context *DeploymentContext) error {
	// delete deployment, pvcs and pvs
	// then
	// create new deployment, pvcs and pvs

	if k8sDeployment := context.Deployment.Subsystems.K8s.GetDeployment(context.Deployment.Name); service.Created(k8sDeployment) {
		err := resources.SsDeleter(context.Client.DeleteDeployment).
			WithResourceID(k8sDeployment.ID).
			WithDbFunc(dbFunc(context.Deployment.ID, "deploymentMap."+context.Deployment.Name)).
			Exec()
		if err != nil {
			return err
		}
	}

	for mapName, pvc := range context.Deployment.Subsystems.K8s.PvcMap {
		err := resources.SsDeleter(context.Client.DeletePVC).
			WithResourceID(pvc.ID).
			WithDbFunc(dbFunc(context.Deployment.ID, "pvcMap."+mapName)).
			Exec()

		if err != nil {
			return err
		}
	}

	for mapName, pv := range context.Deployment.Subsystems.K8s.PvMap {
		err := resources.SsDeleter(context.Client.DeletePV).
			WithResourceID(pv.ID).
			WithDbFunc(dbFunc(context.Deployment.ID, "pvMap."+mapName)).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range context.Generator.PVs() {
		err := resources.SsCreator(context.Client.CreatePV).
			WithDbFunc(dbFunc(context.Deployment.ID, "pvMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range context.Generator.PVCs() {
		err := resources.SsCreator(context.Client.CreatePVC).
			WithDbFunc(dbFunc(context.Deployment.ID, "pvcMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range context.Generator.Deployments() {
		err := resources.SsCreator(context.Client.CreateDeployment).
			WithDbFunc(dbFunc(context.Deployment.ID, "deploymentMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	return nil
}

func updateImage(context *DeploymentContext) error {
	deployments := context.Generator.Deployments()
	idx := slices.IndexFunc(deployments, func(i k8sModels.DeploymentPublic) bool {
		return i.Name == context.Deployment.Name
	})
	if idx == -1 {
		log.Println("main k8s deployment for deployment", context.Deployment.ID, "not found when updating image. assuming it was deleted")
		return nil
	}

	if service.NotCreated(&deployments[idx]) {
		log.Println("main k8s deployment for deployment", context.Deployment.ID, "not created yet when updating image. assuming it was deleted")
		return nil
	}

	err := resources.SsUpdater(context.Client.UpdateDeployment).
		WithDbFunc(dbFunc(context.Deployment.ID, "deploymentMap."+context.Deployment.Name)).
		WithPublic(&deployments[idx]).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		return deploymentModel.New().UpdateSubsystemByID_test(id, "k8s."+key, data)
	}
}
