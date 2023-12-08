package k8s_service

import (
	"context"
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/constants"
	dErrors "go-deploy/service/deployment_service/errors"
	"go-deploy/service/resources"
	"go-deploy/utils"
	"golang.org/x/exp/slices"
	"log"
	"strings"
	"time"
)

func (c *Client) Create(params *deploymentModel.CreateParams) error {
	log.Println("setting up k8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %w", params.Name, err)
	}

	if c.Deployment() == nil {
		return makeError(dErrors.DeploymentNotFoundErr)
	}

	// Namespace
	err := resources.SsCreator(c.Client().CreateNamespace).
		WithDbFunc(dbFunc(c.ID(), "namespace")).
		WithPublic(c.Generator().Namespace()).
		Exec()
	if err != nil {
		return makeError(err)
	}

	// PersistentVolume
	for _, pvPublic := range c.Generator().PVs() {
		err = resources.SsCreator(c.Client().CreatePV).
			WithDbFunc(dbFunc(c.ID(), "pvMap."+pvPublic.Name)).
			WithPublic(&pvPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for _, pvcPublic := range c.Generator().PVCs() {
		err = resources.SsCreator(c.Client().CreatePVC).
			WithDbFunc(dbFunc(c.ID(), "pvcMap."+pvcPublic.Name)).
			WithPublic(&pvcPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for _, secretPublic := range c.Generator().Secrets() {
		err = resources.SsCreator(c.Client().CreateSecret).
			WithDbFunc(dbFunc(c.ID(), "secretMap."+secretPublic.Name)).
			WithPublic(&secretPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deploymentPublic := range c.Generator().Deployments() {
		err = resources.SsCreator(c.Client().CreateDeployment).
			WithDbFunc(dbFunc(c.ID(), "deploymentMap."+deploymentPublic.Name)).
			WithPublic(&deploymentPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, servicePublic := range c.Generator().Services() {
		err = resources.SsCreator(c.Client().CreateService).
			WithDbFunc(dbFunc(c.ID(), "serviceMap."+servicePublic.Name)).
			WithPublic(&servicePublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	for _, ingressPublic := range c.Generator().Ingresses() {
		err = resources.SsCreator(c.Client().CreateIngress).
			WithDbFunc(dbFunc(c.ID(), "ingressMap."+ingressPublic.Name)).
			WithPublic(&ingressPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	for _, hpaPublic := range c.Generator().HPAs() {
		err = resources.SsCreator(c.Client().CreateHPA).
			WithDbFunc(dbFunc(c.ID(), "hpaMap."+hpaPublic.Name)).
			WithPublic(&hpaPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func (c *Client) Delete() error {
	log.Println("deleting k8s for", c.ID())

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %w", c.ID(), err)
	}

	d := c.Deployment()
	if d == nil {
		return makeError(dErrors.DeploymentNotFoundErr)
	}

	kc := c.Client()
	if kc == nil {
		return makeError(dErrors.DeploymentNotFoundErr)
	}

	// Ingress
	for mapName, ingress := range d.Subsystems.K8s.IngressMap {
		err := resources.SsDeleter(kc.DeleteIngress).
			WithResourceID(ingress.ID).
			WithDbFunc(dbFunc(c.ID(), "ingressMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for mapName, k8sService := range d.Subsystems.K8s.ServiceMap {
		err := resources.SsDeleter(kc.DeleteService).
			WithResourceID(k8sService.ID).
			WithDbFunc(dbFunc(c.ID(), "serviceMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for mapName, k8sDeployment := range d.Subsystems.K8s.DeploymentMap {
		err := resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.ID).
			WithDbFunc(dbFunc(c.ID(), "deploymentMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for mapName, pvc := range d.Subsystems.K8s.PvcMap {
		err := resources.SsDeleter(kc.DeletePVC).
			WithResourceID(pvc.ID).
			WithDbFunc(dbFunc(c.ID(), "pvcMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	for mapName, pv := range d.Subsystems.K8s.PvMap {
		err := resources.SsDeleter(kc.DeletePV).
			WithResourceID(pv.ID).
			WithDbFunc(dbFunc(c.ID(), "pvMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Job
	for mapName, job := range d.Subsystems.K8s.JobMap {
		err := resources.SsDeleter(kc.DeleteJob).
			WithResourceID(job.ID).
			WithDbFunc(dbFunc(c.ID(), "jobMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	for mapName, hpa := range d.Subsystems.K8s.HpaMap {
		err := resources.SsDeleter(kc.DeleteHPA).
			WithResourceID(hpa.ID).
			WithDbFunc(dbFunc(c.ID(), "hpaMap."+mapName)).
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
			WithDbFunc(dbFunc(c.ID(), "secretMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Namespace
	err := resources.SsDeleter(func(string) error { return nil }).
		WithResourceID(d.Subsystems.K8s.Namespace.ID).
		WithDbFunc(dbFunc(c.ID(), "namespace")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Update(params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s for deployment %s. details: %w", c.ID(), err)
	}

	if *params == (deploymentModel.UpdateParams{}) {
		return nil
	}

	if params.Name != nil {
		// since names are immutable in k8s, we actually need to recreate everything
		// we can trigger this in a repair.
		// this is rather expensive, but it will include all the other updates as well,
		// so we can just return here
		err := c.Repair()
		if err != nil {
			return makeError(fmt.Errorf("failed to update name in k8s for deployment %s. details: %w", c.ID(), err))
		}

		return nil
	}

	if params.InternalPort != nil {
		err := c.updateInternalPort()
		if err != nil {
			return makeError(err)
		}
	}

	if params.Envs != nil {
		err := c.updateEnvs()
		if err != nil {
			return makeError(err)
		}
	}

	if params.CustomDomain != nil && (params.Private == nil || !*params.Private) {
		err := c.updateCustomDomain()
		if err != nil {
			return makeError(err)
		}
	}

	if params.Private != nil {
		err := c.updatePrivate()
		if err != nil {
			return makeError(err)
		}
	}

	if params.Volumes != nil {
		err := c.recreatePvPvcDeployments()
		if err != nil {
			return makeError(err)
		}
	}

	if params.Image != nil {
		err := c.updateImage()
		if err != nil {
			return makeError(err)
		}
	}

	if params.Replicas != nil {
		err := c.updateReplicas()
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func (c *Client) EnsureOwner(oldOwnerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s owner for deployment %s. details: %w", c.ID(), err)
	}

	if !c.HasID() {
		return makeError(dErrors.DeploymentNotFoundErr)
	}

	// since ownership is determined by the namespace, and the namespace owns everything,
	// we need to recreate everything
	newOwnerID := c.UserID

	// delete everything in the old namespace
	err := c.WithUserID(oldOwnerID).Delete()
	if err != nil {
		return makeError(err)
	}

	// create everything in the new namespace
	err = c.WithUserID(newOwnerID).Repair()
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Restart() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s %s. details: %w", c.ID(), err)
	}

	d := c.Deployment()
	if d == nil {
		return makeError(dErrors.DeploymentNotFoundErr)
	}

	kc := c.Client()
	if kc == nil {
		return makeError(dErrors.DeploymentNotFoundErr)
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

func (c *Client) Repair() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair k8s %s. details: %w", c.ID(), err)
	}

	d := c.Deployment()
	if d == nil {
		return makeError(dErrors.DeploymentNotFoundErr)
	}

	g := c.Generator()
	if g == nil {
		return makeError(dErrors.DeploymentNotFoundErr)
	}

	kc := c.Client()
	if kc == nil {
		return makeError(dErrors.DeploymentNotFoundErr)
	}

	namespace := g.Namespace()
	err := resources.SsRepairer(
		kc.ReadNamespace,
		kc.CreateNamespace,
		kc.UpdateNamespace,
		func(string) error { return nil },
	).WithResourceID(namespace.ID).WithDbFunc(dbFunc(c.ID(), "namespace")).WithGenPublic(namespace).Exec()

	if err != nil {
		return makeError(err)
	}

	deployments := g.Deployments()
	for mapName, k8sDeployment := range d.Subsystems.K8s.GetDeploymentMap() {
		idx := slices.IndexFunc(deployments, func(d k8sModels.DeploymentPublic) bool { return d.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteDeployment).
				WithResourceID(k8sDeployment.ID).
				WithDbFunc(dbFunc(c.ID(), "deploymentMap."+mapName)).
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
		).WithResourceID(public.ID).WithDbFunc(dbFunc(c.ID(), "deploymentMap."+public.Name)).WithGenPublic(&public).Exec()

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
				WithDbFunc(dbFunc(c.ID(), "serviceMap."+mapName)).
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
		).WithResourceID(public.ID).WithDbFunc(dbFunc(c.ID(), "serviceMap."+public.Name)).WithGenPublic(&public).Exec()

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
				WithDbFunc(dbFunc(c.ID(), "ingressMap."+mapName)).
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
		).WithResourceID(public.ID).WithDbFunc(dbFunc(c.ID(), "ingressMap."+public.Name)).WithGenPublic(&public).Exec()
	}

	secrets := g.Secrets()
	for mapName, secret := range d.Subsystems.K8s.GetSecretMap() {
		idx := slices.IndexFunc(secrets, func(s k8sModels.SecretPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteSecret).
				WithResourceID(secret.ID).
				WithDbFunc(dbFunc(c.ID(), "secretMap."+mapName)).
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
		).WithResourceID(public.ID).WithDbFunc(dbFunc(c.ID(), "secretMap."+public.Name)).WithGenPublic(&public).Exec()

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
				WithDbFunc(dbFunc(c.ID(), "hpaMap."+mapName)).
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
		).WithResourceID(public.ID).WithDbFunc(dbFunc(c.ID(), "hpaMap."+public.Name)).WithGenPublic(&public).Exec()

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
				WithDbFunc(dbFunc(c.ID(), "pvcMap."+mapName)).
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
		).WithResourceID(public.ID).WithDbFunc(dbFunc(c.ID(), "pvcMap."+public.Name)).WithGenPublic(&public).Exec()

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
				WithDbFunc(dbFunc(c.ID(), "pvMap."+mapName)).
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
		).WithResourceID(public.ID).WithDbFunc(dbFunc(c.ID(), "pvMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if anyMismatch {
		return c.recreatePvPvcDeployments()
	}

	return nil
}

func (c *Client) SetupLogStream(ctx context.Context, handler func(string, int, time.Time)) error {
	_ = func(err error) error {
		return fmt.Errorf("failed to setup log stream for deployment %s. details: %w", c.ID(), err)
	}

	d := c.Deployment()
	if d == nil {
		return dErrors.DeploymentNotFoundErr
	}

	kc := c.Client()
	if kc == nil {
		return dErrors.DeploymentNotFoundErr
	}

	if d.BeingDeleted() {
		log.Println("deployment", c.ID(), "is being deleted. not setting up log stream")
		return nil
	}

	mainDeployment := d.Subsystems.K8s.GetDeployment(d.Name)
	if !service.Created(mainDeployment) {
		log.Println("main k8s deployment for deployment", c.ID(), "not created when setting up log stream. assuming it was deleted")
		return nil
	}

	err := kc.SetupDeploymentLogStream(ctx, mainDeployment.ID, handler)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) updateInternalPort() error {
	g := c.Generator()
	kc := c.Client()
	d := c.Deployment()

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

	err := resources.SsUpdater(kc.UpdateService).
		WithDbFunc(dbFunc(d.ID, "serviceMap."+d.Name)).
		WithPublic(&services[idx]).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) updateEnvs() error {
	g := c.Generator()
	kc := c.Client()
	d := c.Deployment()

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

	err := resources.SsUpdater(kc.UpdateDeployment).
		WithDbFunc(dbFunc(d.ID, "deploymentMap."+d.Name)).
		WithPublic(&deployments[idx]).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) updateCustomDomain() error {
	g := c.Generator()
	kc := c.Client()
	d := c.Deployment()

	ingresses := g.Ingresses()
	idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool {
		return i.Name == constants.WithCustomDomainSuffix(d.Name)
	})
	if idx == -1 {
		return nil
	}

	var err error

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
		if strings.Contains(err.Error(), "is already defined in ingress") {
			return dErrors.CustomDomainInUseErr
		}

		return err
	}

	return nil
}

func (c *Client) updatePrivate() error {
	g := c.Generator()
	kc := c.Client()
	d := c.Deployment()

	if c.MainApp.Private {
		for mapName, ingress := range d.Subsystems.K8s.IngressMap {
			err := resources.SsDeleter(kc.DeleteIngress).
				WithResourceID(ingress.ID).
				WithDbFunc(dbFunc(d.ID, "ingressMap."+mapName)).
				Exec()

			if err != nil {
				return err
			}
		}
	} else {
		for _, ingressPublic := range g.Ingresses() {
			err := resources.SsCreator(kc.CreateIngress).
				WithDbFunc(dbFunc(d.ID, "ingressMap."+ingressPublic.Name)).
				WithPublic(&ingressPublic).
				Exec()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) updateImage() error {
	g := c.Generator()
	kc := c.Client()
	d := c.Deployment()

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

	err := resources.SsUpdater(kc.UpdateDeployment).
		WithDbFunc(dbFunc(d.ID, "deploymentMap."+d.Name)).
		WithPublic(&deployments[idx]).
		Exec()

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) updateReplicas() error {
	g := c.Generator()
	kc := c.Client()
	d := c.Deployment()

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

	err := resources.SsUpdater(kc.UpdateHPA).
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

func (c *Client) recreatePvPvcDeployments() error {
	// delete deployment, pvcs and pvs
	// then
	// create new deployment, pvcs and pvs

	d := c.Deployment()
	kc := c.Client()
	g := c.Generator()

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

	err := c.Fetch()
	if err != nil {
		if errors.Is(err, dErrors.DeploymentNotFoundErr) {
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

func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return deploymentModel.New().DeleteSubsystemByID(id, "k8s."+key)
		}
		return deploymentModel.New().UpdateSubsystemByID(id, "k8s."+key, data)
	}
}
