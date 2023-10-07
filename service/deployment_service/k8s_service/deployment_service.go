package k8s_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/deployment_service/resources"
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
		WithPublic(context.Generator.Namespace()).
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
	if context.Deployment.Type == deploymentModel.TypeCustom {
		err = resources.SsCreator(context.Client.CreateSecret).
			WithID(deploymentID).
			WithPublic(context.Generator.ImagePullSecret()).
			WithDbKey("k8s.secretMap." + base.AppName).
			Exec()
	}

	// Deployment
	err = resources.SsCreator(context.Client.CreateDeployment).
		WithID(deploymentID).
		WithPublic(context.Generator.MainDeployment()).
		WithDbKey("k8s.deploymentMap." + base.AppName).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Service
	err = resources.SsCreator(context.Client.CreateService).
		WithID(deploymentID).
		WithPublic(context.Generator.MainService()).
		WithDbKey("k8s.serviceMap." + base.AppName).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Ingress main
	var ingressPublic *k8sModels.IngressPublic
	if params.Private {
		ingressPublic = context.Generator.PrivateIngress()
	} else {
		ingressPublic = context.Generator.MainIngress()
	}

	err = resources.SsCreator(context.Client.CreateIngress).
		WithID(deploymentID).
		WithPublic(ingressPublic).
		WithDbKey("k8s.ingressMap." + base.AppName).
		Exec()

	// Ingress custom domain
	if params.CustomDomain != nil && !params.Private {
		err = resources.SsCreator(context.Client.CreateIngress).
			WithID(deploymentID).
			WithPublic(context.Generator.CustomDomainIngress()).
			WithDbKey("k8s.ingressMap." + base.AppNameCustomDomain).
			Exec()
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
		err = resources.SsRepairer(
			context.Client.ReadDeployment,
			context.Client.CreateDeployment,
			context.Client.UpdateDeployment,
			context.Client.DeleteDeployment,
		).WithID(id).WithGenPublicFunc(context.Generator.MainDeployment).WithDbKey("k8s.deploymentMap." + mapName).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	for mapName := range context.Deployment.Subsystems.K8s.ServiceMap {
		err = resources.SsRepairer(
			context.Client.ReadService,
			context.Client.CreateService,
			context.Client.UpdateService,
			context.Client.DeleteService,
		).WithID(id).WithGenPublicFunc(context.Generator.MainService).WithDbKey("k8s.serviceMap." + mapName).Exec()
	}

	if context.MainApp.Private {
		for mapName := range context.Deployment.Subsystems.K8s.IngressMap {
			err = resources.SsDeleter(context.Client.DeleteIngress).
				WithID(id).
				WithDbKey("k8s.ingressMap." + mapName).
				WithResourceID(context.Deployment.Subsystems.K8s.IngressMap[mapName].ID).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	} else {
		if ingress := context.Deployment.Subsystems.K8s.GetIngress(base.AppName); service.Created(ingress) {
			err = resources.SsRepairer(
				context.Client.ReadIngress,
				context.Client.CreateIngress,
				context.Client.UpdateIngress,
				context.Client.DeleteIngress,
			).WithID(id).WithResourceID(ingress.ID).WithGenPublicFunc(context.Generator.MainIngress).WithDbKey("k8s.ingressMap." + base.AppName).Exec()
		} else {
			err = resources.SsCreator(context.Client.CreateIngress).
				WithID(id).
				WithPublic(context.Generator.MainIngress()).
				WithDbKey("k8s.ingressMap." + base.AppName).
				Exec()
		}
		if err != nil {
			return makeError(err)
		}

		if context.MainApp.CustomDomain != nil {
			if ingress := context.Deployment.Subsystems.K8s.GetIngress(base.AppNameCustomDomain); service.Created(ingress) {
				err = resources.SsRepairer(
					context.Client.ReadIngress,
					context.Client.CreateIngress,
					context.Client.UpdateIngress,
					context.Client.DeleteIngress,
				).WithID(id).WithResourceID(ingress.ID).WithGenPublicFunc(context.Generator.CustomDomainIngress).WithDbKey("k8s.ingressMap." + base.AppNameCustomDomain).Exec()
			} else {
				err = resources.SsCreator(context.Client.CreateIngress).
					WithID(id).
					WithPublic(context.Generator.CustomDomainIngress()).
					WithDbKey("k8s.ingressMap." + base.AppNameCustomDomain).
					Exec()
			}
		}
	}

	for mapName := range context.Deployment.Subsystems.K8s.SecretMap {
		err = resources.SsRepairer(
			context.Client.ReadSecret,
			context.Client.CreateSecret,
			context.Client.UpdateSecret,
			context.Client.DeleteSecret,
		).WithID(id).WithGenPublicFunc(context.Generator.ImagePullSecret).WithDbKey("k8s.secretMap." + mapName).Exec()
	}

	return nil
}

func updateInternalPort(context *Context) error {
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

func updateEnvs(context *Context) error {
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

func updateCustomDomain(context *Context) error {
	ingress := context.Deployment.Subsystems.K8s.GetIngress(base.AppNameCustomDomain)
	if service.Created(ingress) {
		ingress.CustomCert = &k8sModels.CustomCert{
			ClusterIssuer: "letsencrypt-prod-deploy-http",
			CommonName:    *context.UpdateParams.CustomDomain,
		}
		ingress.Hosts = []string{*context.UpdateParams.CustomDomain}
	} else {
		ingress = context.Generator.CustomDomainIngress()
	}

	err := resources.SsUpdater(context.Client.UpdateIngress).
		WithID(context.Deployment.ID).
		WithPublic(ingress).
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

func updatePrivate(context *Context) error {
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
		err := resources.SsCreator(context.Client.CreateIngress).
			WithID(context.Deployment.ID).
			WithPublic(context.Generator.MainIngress()).
			WithDbKey("k8s.ingressMap." + base.AppName).
			Exec()
		if err != nil {
			return err
		}

		var customDomain *string
		if context.UpdateParams.CustomDomain != nil {
			customDomain = context.UpdateParams.CustomDomain
		} else {
			customDomain = context.MainApp.CustomDomain
		}

		if customDomain != nil {
			err = resources.SsCreator(context.Client.CreateIngress).
				WithID(context.Deployment.ID).
				WithPublic(context.Generator.CustomDomainIngress()).
				WithDbKey("k8s.ingressMap." + base.AppNameCustomDomain).
				Exec()
		}
	}

	return nil
}

func updateVolumes(context *Context) error {
	// delete deployment, pvcs and pvs
	// then
	// create new deployment, pvcs and pvs
	//
	// this does NOT support multiple k8s deployments yet, as it will delete all then create only the main one

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

	err := resources.SsCreator(context.Client.CreateDeployment).
		WithID(context.Deployment.ID).
		WithPublic(context.Generator.MainDeployment()).
		WithDbKey("k8s.deploymentMap." + base.AppName).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

func updateImage(context *Context) error {
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
