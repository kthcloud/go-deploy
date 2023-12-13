package k8s_service

import (
	"context"
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	kErrors "go-deploy/pkg/subsystems/k8s/errors"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/constants"
	"go-deploy/service/deployment_service/client"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/utils"
	"golang.org/x/exp/slices"
	"log"
	"time"
)

// Create sets up K8s for the deployment.
//
// It creates all necessary resources in K8s, such as namespaces, deployments, services, etc.
func (c *Client) Create(id string, params *deploymentModel.CreateParams) error {
	log.Println("setting up k8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %w", params.Name, err)
	}

	_, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	// Namespace
	err = resources.SsCreator(kc.CreateNamespace).
		WithDbFunc(dbFunc(id, "namespace")).
		WithPublic(g.Namespace()).
		Exec()
	if err != nil {
		return makeError(err)
	}

	// PersistentVolume
	for _, pvPublic := range g.PVs() {
		err = resources.SsCreator(kc.CreatePV).
			WithDbFunc(dbFunc(id, "pvMap."+pvPublic.Name)).
			WithPublic(&pvPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for _, pvcPublic := range g.PVCs() {
		err = resources.SsCreator(kc.CreatePVC).
			WithDbFunc(dbFunc(id, "pvcMap."+pvcPublic.Name)).
			WithPublic(&pvcPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for _, secretPublic := range g.Secrets() {
		err = resources.SsCreator(kc.CreateSecret).
			WithDbFunc(dbFunc(id, "secretMap."+secretPublic.Name)).
			WithPublic(&secretPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deploymentPublic := range g.Deployments() {
		err = resources.SsCreator(kc.CreateDeployment).
			WithDbFunc(dbFunc(id, "deploymentMap."+deploymentPublic.Name)).
			WithPublic(&deploymentPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, servicePublic := range g.Services() {
		err = resources.SsCreator(kc.CreateService).
			WithDbFunc(dbFunc(id, "serviceMap."+servicePublic.Name)).
			WithPublic(&servicePublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	for _, ingressPublic := range g.Ingresses() {
		err = resources.SsCreator(kc.CreateIngress).
			WithDbFunc(dbFunc(id, "ingressMap."+ingressPublic.Name)).
			WithPublic(&ingressPublic).
			Exec()

		if err != nil {
			if errors.Is(err, kErrors.IngressHostInUseErr) {
				return makeError(sErrors.IngressHostInUseErr)
			}

			return makeError(err)
		}
	}

	for _, hpaPublic := range g.HPAs() {
		err = resources.SsCreator(kc.CreateHPA).
			WithDbFunc(dbFunc(id, "hpaMap."+hpaPublic.Name)).
			WithPublic(&hpaPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

// Delete deletes all K8s resources for the deployment.
//
// It deletes all K8s resources, such as namespaces, deployments, services, etc.
func (c *Client) Delete(id string, overwriteUserID ...string) error {
	log.Println("deleting k8s for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %w", id, err)
	}

	var userID string
	if len(overwriteUserID) > 0 {
		userID = overwriteUserID[0]
	}

	d, kc, _, err := c.Get(OptsNoGenerator(id, client.ExtraOpts{UserID: userID}))
	if err != nil {
		return makeError(err)
	}

	// Ingress
	for mapName, ingress := range d.Subsystems.K8s.IngressMap {
		err := resources.SsDeleter(kc.DeleteIngress).
			WithResourceID(ingress.ID).
			WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for mapName, k8sService := range d.Subsystems.K8s.ServiceMap {
		err := resources.SsDeleter(kc.DeleteService).
			WithResourceID(k8sService.ID).
			WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for mapName, k8sDeployment := range d.Subsystems.K8s.DeploymentMap {
		err := resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.ID).
			WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for mapName, pvc := range d.Subsystems.K8s.PvcMap {
		err := resources.SsDeleter(kc.DeletePVC).
			WithResourceID(pvc.ID).
			WithDbFunc(dbFunc(id, "pvcMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	for mapName, pv := range d.Subsystems.K8s.PvMap {
		err := resources.SsDeleter(kc.DeletePV).
			WithResourceID(pv.ID).
			WithDbFunc(dbFunc(id, "pvMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Job
	for mapName, job := range d.Subsystems.K8s.JobMap {
		err := resources.SsDeleter(kc.DeleteJob).
			WithResourceID(job.ID).
			WithDbFunc(dbFunc(id, "jobMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	for mapName, hpa := range d.Subsystems.K8s.HpaMap {
		err := resources.SsDeleter(kc.DeleteHPA).
			WithResourceID(hpa.ID).
			WithDbFunc(dbFunc(id, "hpaMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for mapName, secret := range d.Subsystems.K8s.SecretMap {
		var deleteFunc func(id string) error
		if mapName == constants.WildcardCertSecretName {
			deleteFunc = func(string) error { return nil }
		} else {
			deleteFunc = kc.DeleteSecret
		}

		err := resources.SsDeleter(deleteFunc).
			WithResourceID(secret.ID).
			WithDbFunc(dbFunc(id, "secretMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Namespace
	err = resources.SsDeleter(func(string) error { return nil }).
		WithResourceID(d.Subsystems.K8s.Namespace.ID).
		WithDbFunc(dbFunc(id, "namespace")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

// Update updates K8s resources for the deployment.
//
// It updates all K8s resources tied to the fields in the deployment.UpdateParams.
func (c *Client) Update(id string, params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s for deployment %s. details: %w", id, err)
	}

	if *params == (deploymentModel.UpdateParams{}) {
		return nil
	}

	if params.Name != nil {
		// since names are immutable in k8s, we actually need to recreate everything
		// we can trigger this in a repair.
		// this is rather expensive, but it will include all the other updates as well,
		// so we can just return here
		err := c.Repair(id)
		if err != nil {
			return makeError(fmt.Errorf("failed to update name in k8s for deployment %s. details: %w", id, err))
		}

		return nil
	}

	if params.InternalPort != nil {
		err := c.updateInternalPort(id)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Envs != nil {
		err := c.updateEnvs(id)
		if err != nil {
			return makeError(err)
		}
	}

	if params.CustomDomain != nil && (params.Private == nil || !*params.Private) {
		err := c.updateCustomDomain(id)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Private != nil {
		err := c.updatePrivate(id)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Volumes != nil {
		err := c.recreatePvPvcDeployments(id)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Image != nil {
		err := c.updateImage(id)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Replicas != nil {
		err := c.updateReplicas(id)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

// EnsureOwner ensures the owner of the K8s setup for the deployment.
//
// If the owner of the deployment does match with the ID specified by WithUserID,
// it will update the Harbor setup to match the new owner.
//
// This will always trigger a call to Repair.
func (c *Client) EnsureOwner(id string, oldOwnerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s owner for deployment %s. details: %w", id, err)
	}

	// Since ownership is determined by the namespace, and the namespace owns everything,
	// We need to recreate everything

	// Delete everything related to the deployment in the old namespace
	err := c.Delete(id, oldOwnerID)
	if err != nil {
		return makeError(err)
	}

	// Create everything related to the deployment in the new namespace
	err = c.Repair(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Restart restarts the deployment.
func (c *Client) Restart(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s %s. details: %w", id, err)
	}

	d, kc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		return makeError(err)
	}

	if k8sDeployment := d.Subsystems.K8s.GetDeployment(d.Name); service.Created(k8sDeployment) {
		err := kc.RestartDeployment(k8sDeployment.ID)
		if err != nil {
			return makeError(err)
		}
	} else {
		utils.PrettyPrintError(fmt.Errorf("k8s deployment %s not found when restarting, assuming it was deleted", d.Name))
	}

	return nil
}

// Repair repairs the deployment.
//
// It repairs all K8s resources for the deployment.
func (c *Client) Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair k8s %s. details: %w", id, err)
	}

	d, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	namespace := g.Namespace()
	err = resources.SsRepairer(
		kc.ReadNamespace,
		kc.CreateNamespace,
		kc.UpdateNamespace,
		func(string) error { return nil },
	).WithResourceID(namespace.ID).WithDbFunc(dbFunc(id, "namespace")).WithGenPublic(namespace).Exec()

	if err != nil {
		return makeError(err)
	}

	deployments := g.Deployments()
	for mapName, k8sDeployment := range d.Subsystems.K8s.GetDeploymentMap() {
		idx := slices.IndexFunc(deployments, func(d k8sModels.DeploymentPublic) bool { return d.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteDeployment).
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
			kc.ReadDeployment,
			kc.CreateDeployment,
			kc.UpdateDeployment,
			kc.DeleteDeployment,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "deploymentMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	services := g.Services()
	for mapName, k8sService := range d.Subsystems.K8s.GetServiceMap() {
		idx := slices.IndexFunc(services, func(s k8sModels.ServicePublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteService).
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
			kc.ReadService,
			kc.CreateService,
			kc.UpdateService,
			kc.DeleteService,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "serviceMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	ingresses := g.Ingresses()
	for mapName, ingress := range d.Subsystems.K8s.GetIngressMap() {
		idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool { return i.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteIngress).
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
			kc.ReadIngress,
			kc.CreateIngress,
			kc.UpdateIngress,
			kc.DeleteIngress,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "ingressMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			if errors.Is(err, kErrors.IngressHostInUseErr) {
				return makeError(sErrors.IngressHostInUseErr)
			}

			return makeError(err)
		}
	}

	secrets := g.Secrets()
	for mapName, secret := range d.Subsystems.K8s.GetSecretMap() {
		idx := slices.IndexFunc(secrets, func(s k8sModels.SecretPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteSecret).
				WithResourceID(secret.ID).
				WithDbFunc(dbFunc(id, "secretMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range secrets {
		err = resources.SsRepairer(
			kc.ReadSecret,
			kc.CreateSecret,
			kc.UpdateSecret,
			kc.DeleteSecret,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "secretMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	hpas := g.HPAs()
	for mapName, hpa := range d.Subsystems.K8s.GetHpaMap() {
		idx := slices.IndexFunc(hpas, func(s k8sModels.HpaPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteHPA).
				WithResourceID(hpa.ID).
				WithDbFunc(dbFunc(id, "hpaMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range hpas {
		err = resources.SsRepairer(
			kc.ReadHPA,
			kc.CreateHPA,
			kc.UpdateHPA,
			kc.DeleteHPA,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "hpaMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// the following are special cases because of dependencies between pvcs, pvs and deployments
	// if we have any mismatch for pv or pvc, we need to delete and recreate everything

	anyMismatch := false

	pvcs := g.PVCs()
	for mapName, pvc := range d.Subsystems.K8s.PvcMap {
		idx := slices.IndexFunc(pvcs, func(s k8sModels.PvcPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeletePVC).
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
	for _, public := range g.PVCs() {
		err = resources.SsRepairer(
			kc.ReadPVC,
			kc.CreatePVC,
			func(_ *k8sModels.PvcPublic) (*k8sModels.PvcPublic, error) { anyMismatch = true; return &public, nil },
			func(id string) error { return nil },
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "pvcMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	pvs := g.PVs()
	for mapName, pv := range d.Subsystems.K8s.PvMap {
		idx := slices.IndexFunc(pvs, func(s k8sModels.PvPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeletePV).
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
	for _, public := range g.PVs() {
		err = resources.SsRepairer(
			kc.ReadPV,
			kc.CreatePV,
			func(_ *k8sModels.PvPublic) (*k8sModels.PvPublic, error) { anyMismatch = true; return &public, nil },
			func(id string) error { return nil },
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "pvMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if anyMismatch {
		return c.recreatePvPvcDeployments(id)
	}

	return nil
}

// SetupLogStream sets up a log stream for the deployment.
//
// It sets up a log stream for all the pods in the deployment.
// The handler function is called for each log line.
func (c *Client) SetupLogStream(id string, ctx context.Context, handler func(string, int, time.Time)) error {
	_ = func(err error) error {
		return fmt.Errorf("failed to setup log stream for deployment %s. details: %w", id, err)
	}

	d, kc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		return err
	}

	if d.BeingDeleted() {
		log.Println("deployment", id, "is being deleted. not setting up log stream")
		return nil
	}

	mainDeployment := d.Subsystems.K8s.GetDeployment(d.Name)
	if !service.Created(mainDeployment) {
		log.Println("main k8s deployment for deployment", id, "not created when setting up log stream. assuming it was deleted")
		return nil
	}

	err = kc.SetupDeploymentLogStream(ctx, mainDeployment.ID, handler)
	if err != nil {
		return err
	}

	return nil
}

// updateInternalPort updates the internal port for the deployment.
func (c *Client) updateInternalPort(id string) error {
	d, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return err
	}

	services := g.Services()
	idx := slices.IndexFunc(services, func(i k8sModels.ServicePublic) bool {
		return i.Name == d.Name
	})
	if idx == -1 {
		log.Println("main k8s service for deployment", d.ID, "not found when updating internal port. assuming it was deleted")
		return nil
	}

	if service.NotCreated(&services[idx]) {
		log.Println("main k8s service for deployment", d.ID, "not created yet when updating internal port. assuming it was deleted")
		return nil
	}

	err = resources.SsUpdater(kc.UpdateService).
		WithDbFunc(dbFunc(d.ID, "serviceMap."+d.Name)).
		WithPublic(&services[idx]).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

// updateEnvs updates the envs for the deployment.
func (c *Client) updateEnvs(id string) error {
	d, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return err
	}

	deployments := g.Deployments()
	idx := slices.IndexFunc(deployments, func(i k8sModels.DeploymentPublic) bool {
		return i.Name == d.Name
	})
	if idx == -1 {
		log.Println("main k8s deployment for deployment", d.ID, "not found when updating envs. assuming it was deleted")
		return nil
	}

	if service.NotCreated(&deployments[idx]) {
		log.Println("main k8s deployment for deployment", d.ID, "not created yet when updating envs. assuming it was deleted")
		return nil
	}

	err = resources.SsUpdater(kc.UpdateDeployment).
		WithDbFunc(dbFunc(d.ID, "deploymentMap."+d.Name)).
		WithPublic(&deployments[idx]).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

// updateCustomDomain updates the custom domain for the deployment.
func (c *Client) updateCustomDomain(id string) error {
	d, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return err
	}

	ingresses := g.Ingresses()
	idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool {
		return i.Name == constants.WithCustomDomainSuffix(d.Name)
	})
	if idx == -1 {
		return nil
	}

	if service.NotCreated(&ingresses[idx]) {
		err = resources.SsCreator(kc.CreateIngress).
			WithDbFunc(dbFunc(d.ID, "ingressMap."+constants.WithCustomDomainSuffix(d.Name))).
			WithPublic(&ingresses[idx]).
			Exec()
	} else {
		err = resources.SsUpdater(kc.UpdateIngress).
			WithDbFunc(dbFunc(d.ID, "ingressMap."+constants.WithCustomDomainSuffix(d.Name))).
			WithPublic(&ingresses[idx]).
			Exec()
	}

	if err != nil {
		if errors.Is(err, kErrors.IngressHostInUseErr) {
			return sErrors.IngressHostInUseErr
		}

		return err
	}

	return nil
}

// updatePrivate updates the private for the deployment.
func (c *Client) updatePrivate(id string) error {
	d, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return err
	}

	if d.GetMainApp().Private {
		for mapName, ingress := range d.Subsystems.K8s.IngressMap {
			err = resources.SsDeleter(kc.DeleteIngress).
				WithResourceID(ingress.ID).
				WithDbFunc(dbFunc(d.ID, "ingressMap."+mapName)).
				Exec()

			if err != nil {
				return err
			}
		}
	} else {
		for _, ingressPublic := range g.Ingresses() {
			err = resources.SsCreator(kc.CreateIngress).
				WithDbFunc(dbFunc(d.ID, "ingressMap."+ingressPublic.Name)).
				WithPublic(&ingressPublic).
				Exec()

			if err != nil {
				if errors.Is(err, kErrors.IngressHostInUseErr) {
					return sErrors.IngressHostInUseErr
				}

				return err
			}
		}
	}

	return nil
}

// updateImage updates the image for the deployment.
func (c *Client) updateImage(id string) error {
	d, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return err
	}

	deployments := g.Deployments()
	idx := slices.IndexFunc(deployments, func(i k8sModels.DeploymentPublic) bool {
		return i.Name == d.Name
	})
	if idx == -1 {
		log.Println("main k8s deployment for deployment", d.ID, "not found when updating image. assuming it was deleted")
		return nil
	}

	if service.NotCreated(&deployments[idx]) {
		log.Println("main k8s deployment for deployment", d.ID, "not created yet when updating image. assuming it was deleted")
		return nil
	}

	err = resources.SsUpdater(kc.UpdateDeployment).
		WithDbFunc(dbFunc(d.ID, "deploymentMap."+d.Name)).
		WithPublic(&deployments[idx]).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

// updateReplicas updates the replicas for the deployment.
func (c *Client) updateReplicas(id string) error {
	d, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return err
	}

	hpas := g.HPAs()
	idx := slices.IndexFunc(hpas, func(i k8sModels.HpaPublic) bool {
		return i.Name == d.Name
	})
	if idx == -1 {
		log.Println("main k8s hpa for deployment", d.ID, "not found when updating replicas. assuming it was deleted")
		return nil
	}

	if service.NotCreated(&hpas[idx]) {
		log.Println("main k8s hpa for deployment", d.ID, "not created yet when updating replicas. assuming it was deleted")
		return nil
	}

	deployments := g.Deployments()
	idx = slices.IndexFunc(deployments, func(i k8sModels.DeploymentPublic) bool {
		return i.Name == d.Name
	})
	if idx == -1 {
		log.Println("main k8s deployment for deployment", d.ID, "not found when updating replicas. assuming it was deleted")
		return nil
	}

	if service.NotCreated(&deployments[idx]) {
		log.Println("main k8s deployment for deployment", d.ID, "not created yet when updating replicas. assuming it was deleted")
		return nil
	}

	err = resources.SsUpdater(kc.UpdateHPA).
		WithDbFunc(dbFunc(d.ID, "hpaMap."+d.Name)).
		WithPublic(&hpas[idx]).
		Exec()

	if err != nil {
		return err
	}

	err = resources.SsUpdater(kc.UpdateDeployment).
		WithDbFunc(dbFunc(d.ID, "deploymentMap."+d.Name)).
		WithPublic(&deployments[idx]).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

// recreatePvPvcDeployments recreates the pv, pvc and deployment for the deployment.
//
// This is needed when the PV or PVC are updated, since they are sticky to the deployment.
// They are recreated in the following fashion: Deployment -> PVC -> PV -> PV -> PVC -> Deployment
func (c *Client) recreatePvPvcDeployments(id string) error {
	// delete deployment, pvcs and pvs
	// then
	// create new deployment, pvcs and pvs

	d, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return err
	}

	if k8sDeployment := d.Subsystems.K8s.GetDeployment(d.Name); service.Created(k8sDeployment) {
		err := resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.ID).
			WithDbFunc(dbFunc(d.ID, "deploymentMap."+d.Name)).
			Exec()
		if err != nil {
			return err
		}
	}

	for mapName, pvc := range d.Subsystems.K8s.PvcMap {
		err := resources.SsDeleter(kc.DeletePVC).
			WithResourceID(pvc.ID).
			WithDbFunc(dbFunc(d.ID, "pvcMap."+mapName)).
			Exec()

		if err != nil {
			return err
		}
	}

	for mapName, pv := range d.Subsystems.K8s.PvMap {
		err := resources.SsDeleter(kc.DeletePV).
			WithResourceID(pv.ID).
			WithDbFunc(dbFunc(d.ID, "pvMap."+mapName)).
			Exec()

		if err != nil {
			return err
		}
	}

	_, err = c.Refresh(id)
	if err != nil {
		if errors.Is(err, sErrors.DeploymentNotFoundErr) {
			return nil
		}

		return err
	}

	for _, public := range g.PVs() {
		err = resources.SsCreator(kc.CreatePV).
			WithDbFunc(dbFunc(d.ID, "pvMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range g.PVCs() {
		err = resources.SsCreator(kc.CreatePVC).
			WithDbFunc(dbFunc(d.ID, "pvcMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range g.Deployments() {
		err = resources.SsCreator(kc.CreateDeployment).
			WithDbFunc(dbFunc(d.ID, "deploymentMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	return nil
}

// dbFunc returns a function that updates the K8s subsystem.
func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return deploymentModel.New().DeleteSubsystemByID(id, "k8s."+key)
		}
		return deploymentModel.New().UpdateSubsystemByID(id, "k8s."+key, data)
	}
}
