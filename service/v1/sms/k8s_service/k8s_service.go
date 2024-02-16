package k8s_service

import (
	"errors"
	"fmt"
	smModels "go-deploy/models/sys/sm"
	kErrors "go-deploy/pkg/subsystems/k8s/errors"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"golang.org/x/exp/slices"
	"log"
)

// Create creates the storage manager.
//
// It creates all K8s resources for the storage manager.
func (c *Client) Create(id string, params *smModels.CreateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create storage manager in k8s. details: %w", err)
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

	// Job
	// These are run without saving to the database, as they will be deleted when completed
	for _, jobPublic := range g.Jobs() {
		err = kc.CreateOneShotJob(&jobPublic)
		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for _, secret := range g.Secrets() {
		err = resources.SsCreator(kc.CreateSecret).
			WithDbFunc(dbFunc(id, "secretMap."+secret.Name)).
			WithPublic(&secret).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deployment := range g.Deployments() {
		err = resources.SsCreator(kc.CreateDeployment).
			WithDbFunc(dbFunc(id, "deploymentMap."+deployment.Name)).
			WithPublic(&deployment).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, k8sService := range g.Services() {
		err = resources.SsCreator(kc.CreateService).
			WithDbFunc(dbFunc(id, "serviceMap."+k8sService.Name)).
			WithPublic(&k8sService).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	for _, ingress := range g.Ingresses() {
		err = resources.SsCreator(kc.CreateIngress).
			WithDbFunc(dbFunc(id, "ingressMap."+ingress.Name)).
			WithPublic(&ingress).
			Exec()

		if err != nil {
			if errors.Is(err, kErrors.IngressHostInUseErr) {
				return makeError(sErrors.IngressHostInUseErr)
			}

			return makeError(err)
		}
	}

	return nil
}

// Delete deletes the storage manager.
//
// It deletes all K8s resources for the storage manager.
func (c *Client) Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete storage manager in k8s. details: %w", err)
	}

	log.Println("deleting k8s for storage manager", id)

	sm, kc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		return makeError(err)
	}

	// Deployment
	for mapName, k8sDeployment := range sm.Subsystems.K8s.DeploymentMap {
		err = resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.Name).
			WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
			Exec()
	}

	// Service
	for mapName, k8sService := range sm.Subsystems.K8s.ServiceMap {
		err = resources.SsDeleter(kc.DeleteService).
			WithResourceID(k8sService.Name).
			WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
			Exec()
	}

	// Ingress
	for mapName, ingress := range sm.Subsystems.K8s.IngressMap {
		err = resources.SsDeleter(kc.DeleteIngress).
			WithResourceID(ingress.Name).
			WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
			Exec()
	}

	// PersistentVolumeClaim
	for mapName, pvc := range sm.Subsystems.K8s.PvcMap {
		err = resources.SsDeleter(kc.DeletePVC).
			WithResourceID(pvc.Name).
			WithDbFunc(dbFunc(id, "pvcMap."+mapName)).
			Exec()
	}

	// PersistentVolume
	for mapName, pv := range sm.Subsystems.K8s.PvMap {
		err = resources.SsDeleter(kc.DeletePV).
			WithResourceID(pv.Name).
			WithDbFunc(dbFunc(id, "pvMap."+mapName)).
			Exec()
	}

	// Secret
	for mapName, secret := range sm.Subsystems.K8s.SecretMap {
		var deleteFunc func(interface{}) error
		if mapName == constants.WildcardCertSecretName {
			deleteFunc = func(interface{}) error { return nil }
		} else {
			deleteFunc = dbFunc(id, "secretMap."+mapName)
		}

		err = resources.SsDeleter(kc.DeleteSecret).
			WithResourceID(secret.Name).
			WithDbFunc(deleteFunc).
			Exec()
	}

	return nil
}

// Repair repairs the storage manager.
//
// It repairs all K8s resources for the storage manager.
func (c *Client) Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair storage manager %s in k8s. details: %w", id, err)
	}

	sm, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	namespace := g.Namespace()
	err = resources.SsRepairer(
		kc.ReadNamespace,
		kc.CreateNamespace,
		kc.UpdateNamespace,
		func(string) error { return nil },
	).WithResourceID(namespace.Name).WithDbFunc(dbFunc(id, "namespace")).WithGenPublic(namespace).Exec()

	if err != nil {
		return makeError(err)
	}

	deployments := g.Deployments()
	for mapName, k8sDeployment := range sm.Subsystems.K8s.DeploymentMap {
		idx := slices.IndexFunc(deployments, func(d k8sModels.DeploymentPublic) bool { return d.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteDeployment).
				WithResourceID(k8sDeployment.Name).
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
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "deploymentMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	services := g.Services()
	for mapName, k8sService := range sm.Subsystems.K8s.ServiceMap {
		idx := slices.IndexFunc(services, func(s k8sModels.ServicePublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteService).
				WithResourceID(k8sService.Name).
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
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "serviceMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	ingresses := g.Ingresses()
	for mapName, ingress := range sm.Subsystems.K8s.IngressMap {
		idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool { return i.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteIngress).
				WithResourceID(ingress.Name).
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
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "ingressMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			if errors.Is(err, kErrors.IngressHostInUseErr) {
				return makeError(sErrors.IngressHostInUseErr)
			}

			return makeError(err)
		}
	}

	secrets := g.Secrets()
	for mapName, secret := range sm.Subsystems.K8s.SecretMap {
		idx := slices.IndexFunc(secrets, func(s k8sModels.SecretPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteSecret).
				WithResourceID(secret.Name).
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
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "secretMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

// dbFunc returns a function that updates the K8s subsystem.
func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return smModels.New().DeleteSubsystem(id, "k8s."+key)
		}
		return smModels.New().SetSubsystem(id, "k8s."+key, data)
	}
}
