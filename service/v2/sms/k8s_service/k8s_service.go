package k8s_service

import (
	"errors"
	"fmt"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/sm_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	kErrors "github.com/kthcloud/go-deploy/pkg/subsystems/k8s/errors"
	k8sModels "github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	"github.com/kthcloud/go-deploy/service/constants"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/resources"
	"golang.org/x/exp/slices"
)

// Create creates the storage manager.
//
// It creates all K8s resources for the storage manager.
func (c *Client) Create(id string, params *model.SmCreateParams) error {
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

	// OneShotJob
	// These are run without saving to the database, as they will be deleted when completed
	for _, jobPublic := range g.OneShotJobs() {
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
			if errors.Is(err, kErrors.ErrIngressHostInUse) {
				return makeError(sErrors.ErrIngressHostInUse)
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

	log.Println("Deleting K8s for storage manager", id)

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
		var deleteFunc func(string) error
		if mapName == constants.WildcardCertSecretName {
			deleteFunc = func(string) error { return nil }
		} else {
			deleteFunc = kc.DeleteSecret
		}

		err = resources.SsDeleter(deleteFunc).
			WithResourceID(secret.Name).
			WithDbFunc(dbFunc(id, "secretMap."+mapName)).
			Exec()
		if err != nil {
			return makeError(err)
		}
	}

	// Namespace
	err = resources.SsDeleter(func(string) error { return nil }).
		WithResourceID(sm.Subsystems.K8s.Namespace.Name).
		WithDbFunc(dbFunc(id, "namespace")).
		Exec()
	if err != nil {
		return makeError(err)
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
			if errors.Is(err, kErrors.ErrIngressHostInUse) {
				return makeError(sErrors.ErrIngressHostInUse)
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

	// The following are special cases because of dependencies between pvcs, pvs and deployments.
	// If we have any mismatch for pv or pvc, we need to delete and recreate everything

	anyMismatch := false

	pvcs := g.PVCs()
	for mapName := range sm.Subsystems.K8s.PvcMap {
		idx := slices.IndexFunc(pvcs, func(s k8sModels.PvcPublic) bool { return s.Name == mapName })
		if idx == -1 {
			anyMismatch = true
			break
		}
	}
	for _, public := range pvcs {
		err = resources.SsRepairer(
			kc.ReadPVC,
			kc.CreatePVC,
			func(_ *k8sModels.PvcPublic) (*k8sModels.PvcPublic, error) { anyMismatch = true; return &public, nil },
			func(id string) error { return nil },
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "pvcMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	pvs := g.PVs()
	for mapName := range sm.Subsystems.K8s.PvMap {
		idx := slices.IndexFunc(pvs, func(s k8sModels.PvPublic) bool { return s.Name == mapName })
		if idx == -1 {
			anyMismatch = true
			break
		}
	}
	for _, public := range pvs {
		err = resources.SsRepairer(
			kc.ReadPV,
			kc.CreatePV,
			func(_ *k8sModels.PvPublic) (*k8sModels.PvPublic, error) { anyMismatch = true; return &public, nil },
			func(id string) error { return nil },
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "pvMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if anyMismatch {
		return c.recreatePvPvcDeployments(id)
	}

	return nil
}

// recreatePvPvcDeployments recreates the pv, pvc and deployment for the deployment.
//
// This is needed when the PV or PVC are updated, since they are sticky to the deployment.
// They are recreated in the following fashion: Deployment -> PVC -> PV -> PV -> PVC -> Deployment
func (c *Client) recreatePvPvcDeployments(id string) error {
	// delete deployments, pvcs and pvs
	// then
	// create new deployments, pvcs and pvs

	sm, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return err
	}

	for mapName, k8sDeployment := range sm.Subsystems.K8s.DeploymentMap {
		err := resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.Name).
			WithDbFunc(dbFunc(sm.ID, "deploymentMap."+mapName)).
			Exec()
		if err != nil {
			return err
		}
	}

	for mapName, pvc := range sm.Subsystems.K8s.PvcMap {
		err := resources.SsDeleter(kc.DeletePVC).
			WithResourceID(pvc.Name).
			WithDbFunc(dbFunc(sm.ID, "pvcMap."+mapName)).
			Exec()

		if err != nil {
			return err
		}
	}

	for mapName, pv := range sm.Subsystems.K8s.PvMap {
		err := resources.SsDeleter(kc.DeletePV).
			WithResourceID(pv.Name).
			WithDbFunc(dbFunc(sm.ID, "pvMap."+mapName)).
			Exec()

		if err != nil {
			return err
		}
	}

	sm, err = c.Refresh(id)
	if err != nil {
		if errors.Is(err, sErrors.ErrSmNotFound) {
			return nil
		}

		return err
	}

	for _, public := range g.PVs() {
		err = resources.SsCreator(kc.CreatePV).
			WithDbFunc(dbFunc(sm.ID, "pvMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range g.PVCs() {
		err = resources.SsCreator(kc.CreatePVC).
			WithDbFunc(dbFunc(sm.ID, "pvcMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, public := range g.Deployments() {
		err = resources.SsCreator(kc.CreateDeployment).
			WithDbFunc(dbFunc(sm.ID, "deploymentMap."+public.Name)).
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
			return sm_repo.New().DeleteSubsystem(id, "k8s."+key)
		}
		return sm_repo.New().SetSubsystem(id, "k8s."+key, data)
	}
}
