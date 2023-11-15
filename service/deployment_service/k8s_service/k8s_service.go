package k8s_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/constants"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/resources"
	"go-deploy/utils"
	"golang.org/x/exp/slices"
	"log"
	"strings"
)

func Create(id string, params *deploymentModel.CreateParams) error {
	log.Println("setting up k8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %w", params.Name, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	// Namespace
	err = resources.SsCreator(context.Client.CreateNamespace).
		WithDbFunc(dbFunc(id, "namespace")).
		WithPublic(context.Generator.Namespace()).
		Exec()
	if err != nil {
		return makeError(err)
	}

	// PersistentVolume
	for _, pvPublic := range context.Generator.PVs() {
		err = resources.SsCreator(context.Client.CreatePV).
			WithDbFunc(dbFunc(id, "pvMap."+pvPublic.Name)).
			WithPublic(&pvPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for _, pvcPublic := range context.Generator.PVCs() {
		err = resources.SsCreator(context.Client.CreatePVC).
			WithDbFunc(dbFunc(id, "pvcMap."+pvcPublic.Name)).
			WithPublic(&pvcPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for _, secretPublic := range context.Generator.Secrets() {
		err = resources.SsCreator(context.Client.CreateSecret).
			WithDbFunc(dbFunc(id, "secretMap."+secretPublic.Name)).
			WithPublic(&secretPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deploymentPublic := range context.Generator.Deployments() {
		err = resources.SsCreator(context.Client.CreateDeployment).
			WithDbFunc(dbFunc(id, "deploymentMap."+deploymentPublic.Name)).
			WithPublic(&deploymentPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, servicePublic := range context.Generator.Services() {
		err = resources.SsCreator(context.Client.CreateService).
			WithDbFunc(dbFunc(id, "serviceMap."+servicePublic.Name)).
			WithPublic(&servicePublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	for _, ingressPublic := range context.Generator.Ingresses() {
		err = resources.SsCreator(context.Client.CreateIngress).
			WithDbFunc(dbFunc(id, "ingressMap."+ingressPublic.Name)).
			WithPublic(&ingressPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func Delete(id string, overrideOwnerID ...string) error {
	log.Println("deleting k8s for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %w", id, err)
	}

	context, err := NewContext(id, overrideOwnerID...)
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
		var deleteFunc func(id string) error
		if mapName == constants.WildcardCertSecretName {
			deleteFunc = func(string) error { return nil }
		} else {
			deleteFunc = context.Client.DeleteSecret
		}

		err = resources.SsDeleter(deleteFunc).
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

	if params.Name != nil {
		// since names are immutable in k8s, we actually need to recreate everything
		// we can trigger this in a repair.
		// this is rather expensive, but it will include all the other updates as well,
		// so we can just return here
		err = Repair(id)
		if err != nil {
			return makeError(fmt.Errorf("failed to update name in k8s for deployment %s. details: %w", id, err))
		}

		return nil
	}

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
		err = recreatePvPvcDeployments(context)
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

func EnsureOwner(id, oldOwnerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s owner for deployment %s. details: %w", id, err)
	}

	// since ownership is determined by the namespace, and the namespace owns everything,
	// we need to recreate everything

	// delete everything in the old namespace
	err := Delete(id, oldOwnerID)
	if err != nil {
		return makeError(err)
	}

	// create everything in the new namespace
	err = Repair(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Restart(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s %s. details: %w", id, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	if k8sDeployment := context.Deployment.Subsystems.K8s.GetDeployment(context.Deployment.Name); service.Created(k8sDeployment) {
		err = context.Client.RestartDeployment(k8sDeployment.ID)
		if err != nil {
			return makeError(err)
		}
	} else {
		utils.PrettyPrintError(fmt.Errorf("k8s deployment %s not found when restarting, assuming it was deleted", context.Deployment.Name))
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

	namespace := context.Generator.Namespace()
	err = resources.SsRepairer(
		context.Client.ReadNamespace,
		context.Client.CreateNamespace,
		context.Client.UpdateNamespace,
		func(string) error { return nil },
	).WithResourceID(namespace.ID).WithDbFunc(dbFunc(id, "namespace")).WithGenPublic(namespace).Exec()

	if err != nil {
		return makeError(err)
	}

	deployments := context.Generator.Deployments()
	for mapName, k8sDeployment := range context.Deployment.Subsystems.K8s.DeploymentMap {
		idx := slices.IndexFunc(deployments, func(d k8sModels.DeploymentPublic) bool { return d.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteDeployment).
				WithResourceID(k8sDeployment.ID).
				WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range deployments {
		err = resources.SsRepairer(
			context.Client.ReadDeployment,
			context.Client.CreateDeployment,
			context.Client.UpdateDeployment,
			context.Client.DeleteDeployment,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "deploymentMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	services := context.Generator.Services()
	for mapName, k8sService := range context.Deployment.Subsystems.K8s.ServiceMap {
		idx := slices.IndexFunc(services, func(s k8sModels.ServicePublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteService).
				WithResourceID(k8sService.ID).
				WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range services {
		err = resources.SsRepairer(
			context.Client.ReadService,
			context.Client.CreateService,
			context.Client.UpdateService,
			context.Client.DeleteService,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "serviceMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	ingresses := context.Generator.Ingresses()
	for mapName, ingress := range context.Deployment.Subsystems.K8s.IngressMap {
		idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool { return i.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteIngress).
				WithResourceID(ingress.ID).
				WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range ingresses {
		err = resources.SsRepairer(
			context.Client.ReadIngress,
			context.Client.CreateIngress,
			context.Client.UpdateIngress,
			context.Client.DeleteIngress,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "ingressMap."+public.Name)).WithGenPublic(&public).Exec()
	}

	for mapName, secret := range context.Deployment.Subsystems.K8s.SecretMap {
		secrets := context.Generator.Secrets()
		idx := slices.IndexFunc(secrets, func(s k8sModels.SecretPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeleteSecret).
				WithResourceID(secret.ID).
				WithDbFunc(dbFunc(id, "secretMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range context.Generator.Secrets() {
		err = resources.SsRepairer(
			context.Client.ReadSecret,
			context.Client.CreateSecret,
			context.Client.UpdateSecret,
			context.Client.DeleteSecret,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "secretMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// the following are special cases because of dependencies between pvcs, pvs and deployments
	// if we have any mismatch for pv or pvc, we need to delete and recreate everything

	anyMismatch := false

	pvcs := context.Generator.PVCs()
	for mapName, pvc := range context.Deployment.Subsystems.K8s.PvcMap {
		idx := slices.IndexFunc(pvcs, func(s k8sModels.PvcPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeletePVC).
				WithResourceID(pvc.ID).
				WithDbFunc(dbFunc(id, "pvcMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}

			anyMismatch = true
		}

		if anyMismatch {
			break
		}
	}
	for _, public := range context.Generator.PVCs() {
		err = resources.SsRepairer(
			context.Client.ReadPVC,
			context.Client.CreatePVC,
			func(_ *k8sModels.PvcPublic) (*k8sModels.PvcPublic, error) { anyMismatch = true; return &public, nil },
			func(id string) error { return nil },
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "pvcMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	pvs := context.Generator.PVs()
	for mapName, pv := range context.Deployment.Subsystems.K8s.PvMap {
		idx := slices.IndexFunc(pvs, func(s k8sModels.PvPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeletePV).
				WithResourceID(pv.ID).
				WithDbFunc(dbFunc(id, "pvMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}

			anyMismatch = true
		}

		if anyMismatch {
			break
		}
	}
	for _, public := range context.Generator.PVs() {
		err = resources.SsRepairer(
			context.Client.ReadPV,
			context.Client.CreatePV,
			func(_ *k8sModels.PvPublic) (*k8sModels.PvPublic, error) { anyMismatch = true; return &public, nil },
			func(id string) error { return nil },
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "pvMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if anyMismatch {
		return recreatePvPvcDeployments(context)
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
		return i.Name == constants.WithCustomDomainSuffix(context.Deployment.Name)
	})
	if idx == -1 {
		return nil
	}

	var err error

	if service.NotCreated(&ingresses[idx]) {
		err = resources.SsCreator(context.Client.CreateIngress).
			WithDbFunc(dbFunc(context.Deployment.ID, "ingressMap."+constants.WithCustomDomainSuffix(context.Deployment.Name))).
			WithPublic(&ingresses[idx]).
			Exec()
	} else {
		err = resources.SsUpdater(context.Client.UpdateIngress).
			WithDbFunc(dbFunc(context.Deployment.ID, "ingressMap."+constants.WithCustomDomainSuffix(context.Deployment.Name))).
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

func recreatePvPvcDeployments(context *DeploymentContext) error {
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

	err := context.Refresh()
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return err
	}

	for _, public := range context.Generator.PVs() {
		err = resources.SsCreator(context.Client.CreatePV).
			WithDbFunc(dbFunc(context.Deployment.ID, "pvMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range context.Generator.PVCs() {
		err = resources.SsCreator(context.Client.CreatePVC).
			WithDbFunc(dbFunc(context.Deployment.ID, "pvcMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range context.Generator.Deployments() {
		err = resources.SsCreator(context.Client.CreateDeployment).
			WithDbFunc(dbFunc(context.Deployment.ID, "deploymentMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	return nil
}

func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return deploymentModel.New().DeleteSubsystemByID(id, "k8s."+key)
		}
		return deploymentModel.New().UpdateSubsystemByID(id, "k8s."+key, data)
	}
}