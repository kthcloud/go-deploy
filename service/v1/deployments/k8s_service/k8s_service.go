package k8s_service

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/models/config"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems"
	kErrors "go-deploy/pkg/subsystems/k8s/errors"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/v1/deployments/opts"
	"go-deploy/utils"
	"slices"
	"time"
)

// Create sets up K8s for the deployment.
//
// It creates all necessary resources in K8s, such as namespaces, deployments, services, etc.
func (c *Client) Create(id string, params *model.DeploymentCreateParams) error {
	log.Println("Setting up K8s for", params.Name)

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

	// NetworkPolicies
	for _, networkPolicyPublic := range g.NetworkPolicies() {
		err = resources.SsCreator(kc.CreateNetworkPolicy).
			WithDbFunc(dbFunc(id, "networkPolicyMap."+networkPolicyPublic.Name)).
			WithPublic(&networkPolicyPublic).
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

	// HPA
	for _, hpaPublic := range g.HPAs() {
		err = resources.SsCreator(kc.CreateHPA).
			WithDbFunc(dbFunc(id, "hpaMap."+hpaPublic.Name)).
			WithPublic(&hpaPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// OneShotJobs
	for _, jobPublic := range g.OneShotJobs() {
		err = kc.CreateOneShotJob(&jobPublic)
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
	log.Println("Deleting K8s for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %w", id, err)
	}

	var userID string
	if len(overwriteUserID) > 0 {
		userID = overwriteUserID[0]
	}

	d, kc, _, err := c.Get(OptsNoGenerator(id, opts.ExtraOpts{UserID: userID}))
	if err != nil {
		return makeError(err)
	}

	// Ingress
	for mapName, ingress := range d.Subsystems.K8s.IngressMap {
		err := resources.SsDeleter(kc.DeleteIngress).
			WithResourceID(ingress.Name).
			WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for mapName, k8sService := range d.Subsystems.K8s.ServiceMap {
		err := resources.SsDeleter(kc.DeleteService).
			WithResourceID(k8sService.Name).
			WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for mapName, k8sDeployment := range d.Subsystems.K8s.DeploymentMap {
		err := resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.Name).
			WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for mapName, pvc := range d.Subsystems.K8s.PvcMap {
		err := resources.SsDeleter(kc.DeletePVC).
			WithResourceID(pvc.Name).
			WithDbFunc(dbFunc(id, "pvcMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	for mapName, pv := range d.Subsystems.K8s.PvMap {
		err := resources.SsDeleter(kc.DeletePV).
			WithResourceID(pv.Name).
			WithDbFunc(dbFunc(id, "pvMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	for mapName, hpa := range d.Subsystems.K8s.HpaMap {
		err := resources.SsDeleter(kc.DeleteHPA).
			WithResourceID(hpa.Name).
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
			WithResourceID(secret.Name).
			WithDbFunc(dbFunc(id, "secretMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// NetworkPolicies
	for mapName, networkPolicy := range d.Subsystems.K8s.NetworkPolicyMap {
		err := resources.SsDeleter(kc.DeleteNetworkPolicy).
			WithResourceID(networkPolicy.Name).
			WithDbFunc(dbFunc(id, "networkPolicyMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Namespace
	// They are not deleted in K8s, as they are shared among deployments
	err = resources.SsDeleter(func(string) error { return nil }).
		WithResourceID(d.Subsystems.K8s.Namespace.Name).
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
func (c *Client) Update(id string, params *model.DeploymentUpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s for deployment %s. details: %w", id, err)
	}

	if *params == (model.DeploymentUpdateParams{}) {
		return nil
	}

	_, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	// Since K8s is immutable in many cases, we might need to recreate some resources
	// This logic is already implemented in Repair, so we can just call that
	err = c.Repair(id)
	if err != nil {
		return makeError(err)
	}

	// OneShotJobs
	for _, jobPublic := range g.OneShotJobs() {
		err = kc.CreateOneShotJob(&jobPublic)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

// EnsureOwner ensures the owner of the K8s setup, by deleting and then trigger a call to Repair.
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

	if k8sDeployment := d.Subsystems.K8s.GetDeployment(d.Name); subsystems.Created(k8sDeployment) {
		err := kc.RestartDeployment(k8sDeployment.Name)
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
	).WithResourceID(namespace.Name).WithDbFunc(dbFunc(id, "namespace")).WithGenPublic(namespace).Exec()

	if err != nil {
		return makeError(err)
	}

	networkPolicies := g.NetworkPolicies()
	for mapName, networkPolicy := range d.Subsystems.K8s.GetNetworkPolicyMap() {
		idx := slices.IndexFunc(networkPolicies, func(n k8sModels.NetworkPolicyPublic) bool { return n.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteNetworkPolicy).
				WithResourceID(networkPolicy.Name).
				WithDbFunc(dbFunc(id, "networkPolicyMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range networkPolicies {
		err = resources.SsRepairer(
			kc.ReadNetworkPolicy,
			kc.CreateNetworkPolicy,
			kc.UpdateNetworkPolicy,
			kc.DeleteNetworkPolicy,
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "networkPolicyMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	deployments := g.Deployments()
	for mapName, k8sDeployment := range d.Subsystems.K8s.GetDeploymentMap() {
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
	for mapName, k8sService := range d.Subsystems.K8s.GetServiceMap() {
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
	for mapName, ingress := range d.Subsystems.K8s.GetIngressMap() {
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
	for mapName, secret := range d.Subsystems.K8s.GetSecretMap() {
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

	hpas := g.HPAs()
	for mapName, hpa := range d.Subsystems.K8s.GetHpaMap() {
		idx := slices.IndexFunc(hpas, func(s k8sModels.HpaPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteHPA).
				WithResourceID(hpa.Name).
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
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "hpaMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// The following are special cases because of dependencies between PVCs, PVs and deployments.
	// If we have any mismatch for PV or PVC, we need to delete and recreate everything

	anyMismatch := false

	pvcs := g.PVCs()
	for mapName, _ := range d.Subsystems.K8s.PvcMap {
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
	for mapName, _ := range d.Subsystems.K8s.PvMap {
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

// SetupLogStream sets up a log stream for the deployment.
//
// It sets up a log stream for all the pods in the deployment.
// The handler function is called for each log line.
func (c *Client) SetupLogStream(ctx context.Context, zone *config.Zone, allowedNames []string, handler func(string, string, int, time.Time)) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup log stream for zone %s. details: %w", zone.Name, err)
	}

	_, kc, _, err := c.Get(OptsOnlyClient(zone))
	if err != nil {
		return makeError(err)
	}

	err = kc.SetupLogStream(ctx, allowedNames, handler)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// SetupStatusWatcher sets up a status watcher for a zone.
// For every status change, it triggers the callback.
func (c *Client) SetupStatusWatcher(ctx context.Context, zone *configModels.Zone, resourceType string, callback func(string, interface{})) error {
	_, kc, _, err := c.Get(OptsOnlyClient(zone))
	if err != nil {
		return err
	}

	return kc.SetupWatcher(ctx, resourceType, callback)
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

	if k8sDeployment := d.Subsystems.K8s.GetDeployment(d.Name); subsystems.Created(k8sDeployment) {
		err := resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.Name).
			WithDbFunc(dbFunc(d.ID, "deploymentMap."+d.Name)).
			Exec()
		if err != nil {
			return err
		}
	}

	for mapName, pvc := range d.Subsystems.K8s.PvcMap {
		err := resources.SsDeleter(kc.DeletePVC).
			WithResourceID(pvc.Name).
			WithDbFunc(dbFunc(d.ID, "pvcMap."+mapName)).
			Exec()

		if err != nil {
			return err
		}
	}

	for mapName, pv := range d.Subsystems.K8s.PvMap {
		err := resources.SsDeleter(kc.DeletePV).
			WithResourceID(pv.Name).
			WithDbFunc(dbFunc(d.ID, "pvMap."+mapName)).
			Exec()

		if err != nil {
			return err
		}
	}

	d, err = c.Refresh(id)
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
			return deployment_repo.New().DeleteSubsystem(id, "k8s."+key)
		}
		return deployment_repo.New().SetSubsystem(id, "k8s."+key, data)
	}
}
