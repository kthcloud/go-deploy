package k8s_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/deployment_service/resources"
	"golang.org/x/exp/slices"
	"log"
	"strconv"
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
		WithID(deploymentID).
		WithPublic(context.Generator.MainNamespace()).
		WithDbKey("k8s.namespace").
		Exec()
	if err != nil {
		return makeError(err)
	}

	// PersistentVolume
	for _, pvPublic := range context.Generator.PVs() {
		err = resources.SsCreator(context.Client.CreatePV).
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
			WithID(deploymentID).
			WithPublic(&pvcPublic).
			WithDbKey("k8s.pvcMap." + pvcPublic.Name).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for _, secretPublic := range context.Generator.Secrets() {
		err = resources.SsCreator(context.Client.CreateSecret).
			WithID(deploymentID).
			WithPublic(&secretPublic).
			WithDbKey("k8s.secretMap." + secretPublic.Name).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deploymentPublic := range context.Generator.Deployments() {
		err = resources.SsCreator(context.Client.CreateDeployment).
			WithID(deploymentID).
			WithPublic(&deploymentPublic).
			WithDbKey("k8s.deploymentMap." + deploymentPublic.Name).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, servicePublic := range context.Generator.Services() {
		err = resources.SsCreator(context.Client.CreateService).
			WithID(deploymentID).
			WithPublic(&servicePublic).
			WithDbKey("k8s.serviceMap." + servicePublic.Name).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress main
	for _, ingressPublic := range context.Generator.Ingresses() {
		err = resources.SsCreator(context.Client.CreateIngress).
			WithID(deploymentID).
			WithPublic(&ingressPublic).
			WithDbKey("k8s.ingressMap." + ingressPublic.Name).
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
			WithID(id).
			WithDbKey("k8s.ingressMap." + mapName).
			WithResourceID(ingress.ID).
			Exec()
	}

	// Service
	for mapName, k8sService := range context.Deployment.Subsystems.K8s.ServiceMap {
		err = resources.SsDeleter(context.Client.DeleteService).
			WithID(id).
			WithDbKey("k8s.serviceMap." + mapName).
			WithResourceID(k8sService.ID).
			Exec()
	}

	// Deployment
	for mapName, k8sDeployment := range context.Deployment.Subsystems.K8s.DeploymentMap {
		err = resources.SsDeleter(context.Client.DeleteDeployment).
			WithID(id).
			WithDbKey("k8s.deploymentMap." + mapName).
			WithResourceID(k8sDeployment.ID).
			Exec()
	}

	// PersistentVolumeClaim
	for mapName, pvc := range context.Deployment.Subsystems.K8s.PvcMap {
		err = resources.SsDeleter(context.Client.DeletePVC).
			WithID(id).
			WithDbKey("k8s.pvcMap." + mapName).
			WithResourceID(pvc.ID).
			Exec()
	}

	// PersistentVolume
	for mapName, pv := range context.Deployment.Subsystems.K8s.PvMap {
		err = resources.SsDeleter(context.Client.DeletePV).
			WithID(id).
			WithDbKey("k8s.pvMap." + mapName).
			WithResourceID(pv.ID).
			Exec()
	}

	// Job
	for mapName, job := range context.Deployment.Subsystems.K8s.JobMap {
		err = resources.SsDeleter(context.Client.DeleteJob).
			WithID(id).
			WithDbKey("k8s.jobMap." + mapName).
			WithResourceID(job.ID).
			Exec()
	}

	// Secret
	for mapName, secret := range context.Deployment.Subsystems.K8s.SecretMap {
		err = resources.SsDeleter(context.Client.DeleteSecret).
			WithID(id).
			WithDbKey("k8s.secretMap." + mapName).
			WithResourceID(secret.ID).
			Exec()
	}

	// Namespace
	err = resources.SsDeleter(func(string) error { return nil }).
		WithID(id).
		WithDbKey("k8s.namespace").
		WithResourceID(context.Deployment.Subsystems.K8s.Namespace.ID).
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

	err = resources.SsRepairer(
		context.Client.ReadNamespace,
		context.Client.CreateNamespace,
		context.Client.UpdateNamespace,
		func(string) error { return nil },
	).WithID(id).WithDbKey("k8s.namespace").Exec()

	if err != nil {
		return makeError(err)
	}

	for mapName := range context.Deployment.Subsystems.K8s.DeploymentMap {
		deployments := context.Generator.Deployments()
		idx := slices.IndexFunc(deployments, func(d k8sModels.DeploymentPublic) bool { return d.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteDeployment).
				WithID(id).
				WithDbKey("k8s.deploymentMap." + mapName).
				WithResourceID(context.Deployment.Subsystems.K8s.DeploymentMap[mapName].ID).
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

	for mapName := range context.Deployment.Subsystems.K8s.ServiceMap {
		services := context.Generator.Services()
		idx := slices.IndexFunc(services, func(s k8sModels.ServicePublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteService).
				WithID(id).
				WithDbKey("k8s.serviceMap." + mapName).
				WithResourceID(context.Deployment.Subsystems.K8s.ServiceMap[mapName].ID).
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
	}

	for mapName := range context.Deployment.Subsystems.K8s.IngressMap {
		ingresses := context.Generator.Ingresses()
		idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool { return i.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteIngress).
				WithID(id).
				WithDbKey("k8s.ingressMap." + mapName).
				WithResourceID(context.Deployment.Subsystems.K8s.IngressMap[mapName].ID).
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
	}

	for mapName := range context.Deployment.Subsystems.K8s.SecretMap {
		secrets := context.Generator.Secrets()
		idx := slices.IndexFunc(secrets, func(s k8sModels.SecretPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteSecret).
				WithID(id).
				WithDbKey("k8s.secretMap." + mapName).
				WithResourceID(context.Deployment.Subsystems.K8s.SecretMap[mapName].ID).
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
		).WithID(id).WithGenPublic(&secrets[idx]).WithDbKey("k8s.secretMap." + mapName).Exec()
	}

	return nil
}

func updateInternalPort(context *DeploymentContext) error {
	if k8sService := context.Deployment.Subsystems.K8s.GetService(base.AppName); service.Created(k8sService) {
		if k8sService.Port != *context.UpdateParams.InternalPort {
			k8sService.TargetPort = *context.UpdateParams.InternalPort

			err := resources.SsUpdater(context.Client.UpdateService).
				WithID(context.Deployment.ID).
				WithPublic(k8sService).
				WithDbKey("k8s.serviceMap." + base.AppName).
				Exec()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func updateEnvs(context *DeploymentContext) error {
	if k8sDeployment := context.Deployment.Subsystems.K8s.GetDeployment(base.AppName); service.Created(k8sDeployment) {
		var port int
		if context.UpdateParams.InternalPort != nil {
			port = *context.UpdateParams.InternalPort
		} else {
			port = context.MainApp.InternalPort
		}

		k8sEnvs := []k8sModels.EnvVar{
			{Name: "PORT", Value: strconv.Itoa(port)},
		}

		for _, env := range *context.UpdateParams.Envs {
			if env.Name == "PORT" {
				continue
			}

			k8sEnvs = append(k8sEnvs, k8sModels.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			})
		}

		k8sDeployment.EnvVars = k8sEnvs

		err := resources.SsUpdater(context.Client.UpdateDeployment).
			WithID(context.Deployment.ID).
			WithPublic(k8sDeployment).
			WithDbKey("k8s.deploymentMap." + base.AppName).
			Exec()

		if err != nil {
			return err
		}
	}

	return nil
}

func updateCustomDomain(context *DeploymentContext) error {
	ingresses := context.Generator.Ingresses()
	idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool { return i.Name == base.AppNameCustomDomain })
	if idx == -1 {
		return nil
	}

	err := resources.SsUpdater(context.Client.UpdateIngress).
		WithID(context.Deployment.ID).
		WithPublic(&ingresses[idx]).
		WithDbKey("k8s.ingressMap." + base.AppNameCustomDomain).
		Exec()

	if err != nil {
		if strings.Contains(err.Error(), "is already defined in ingress") {
			return base.CustomDomainInUseErr
		}

		return err
	}

	return nil
}

func updatePrivate(context *DeploymentContext) error {
	if *context.UpdateParams.Private {
		for mapName, ingress := range context.Deployment.Subsystems.K8s.IngressMap {
			err := resources.SsDeleter(context.Client.DeleteIngress).
				WithID(context.Deployment.ID).
				WithDbKey("k8s.ingressMap." + mapName).
				WithResourceID(ingress.ID).
				Exec()

			if err != nil {
				return err
			}
		}
	} else {
		for _, ingressPublic := range context.Generator.Ingresses() {
			err := resources.SsCreator(context.Client.CreateIngress).
				WithID(context.Deployment.ID).
				WithPublic(&ingressPublic).
				WithDbKey("k8s.ingressMap." + ingressPublic.Name).
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

	if k8sDeployment := context.Deployment.Subsystems.K8s.GetDeployment(base.AppName); service.Created(k8sDeployment) {
		err := resources.SsDeleter(context.Client.DeleteDeployment).
			WithID(context.Deployment.ID).
			WithDbKey("k8s.deploymentMap." + base.AppName).
			WithResourceID(k8sDeployment.ID).
			Exec()
		if err != nil {
			return err
		}
	}

	for mapName, pvc := range context.Deployment.Subsystems.K8s.PvcMap {
		err := resources.SsDeleter(context.Client.DeletePVC).
			WithID(context.Deployment.ID).
			WithDbKey("k8s.pvcMap." + mapName).
			WithResourceID(pvc.ID).
			Exec()

		if err != nil {
			return err
		}
	}

	for mapName, pv := range context.Deployment.Subsystems.K8s.PvMap {
		err := resources.SsDeleter(context.Client.DeletePV).
			WithID(context.Deployment.ID).
			WithDbKey("k8s.pvMap." + mapName).
			WithResourceID(pv.ID).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range context.Generator.PVs() {
		err := resources.SsCreator(context.Client.CreatePV).
			WithID(context.Deployment.ID).
			WithDbKey("k8s.pvMap." + public.Name).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range context.Generator.PVCs() {
		err := resources.SsCreator(context.Client.CreatePVC).
			WithID(context.Deployment.ID).
			WithPublic(&public).
			WithDbKey("k8s.pvcMap." + public.Name).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range context.Generator.Deployments() {
		err := resources.SsCreator(context.Client.CreateDeployment).
			WithID(context.Deployment.ID).
			WithPublic(&public).
			WithDbKey("k8s.deploymentMap." + public.Name).
			Exec()

		if err != nil {
			return err
		}
	}

	return nil
}

func updateImage(context *DeploymentContext) error {
	if *context.UpdateParams.Image != context.MainApp.Image {
		oldPublic := context.Deployment.Subsystems.K8s.GetDeployment(base.AppName)
		if oldPublic.Created() {
			newPublic := oldPublic
			newPublic.Image = *context.UpdateParams.Image

			err := resources.SsUpdater(context.Client.UpdateDeployment).
				WithID(context.Deployment.ID).
				WithPublic(newPublic).
				WithDbKey("k8s.deploymentMap." + base.AppName).
				Exec()

			if err != nil {
				return err
			}
		}
	}

	return nil
}
