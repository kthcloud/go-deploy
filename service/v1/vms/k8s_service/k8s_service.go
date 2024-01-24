package k8s_service

import (
	"errors"
	"fmt"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/models/versions"
	kErrors "go-deploy/pkg/subsystems/k8s/errors"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/v1/vms/opts"
	"golang.org/x/exp/slices"
	"log"
)

// Create sets up K8s for a VM.
//
// This does nothing if the VM is version 1 and has no proxy ports
func (c *Client) Create(id string, params *vmModels.CreateParams) error {
	log.Println("setting up k8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %w", params.Name, err)
	}

	_, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			log.Println("vm not found when setting up k8s for", params.Name, ". assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	// Namespace
	namespace := g.Namespace()
	if namespace != nil {
		err = resources.SsCreator(kc.CreateNamespace).
			WithDbFunc(dbFunc(id, "namespace")).
			WithPublic(namespace).
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

	return nil
}

// Delete deletes the K8s setup for a VM.
func (c *Client) Delete(id string, overwriteUserID ...string) error {
	log.Println("deleting k8s for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %w", id, err)
	}

	var userID string
	if len(overwriteUserID) > 0 {
		userID = overwriteUserID[0]
	}

	vm, kc, _, err := c.Get(OptsNoGenerator(id, opts.ExtraOpts{UserID: userID}))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			log.Println("vm not found when deleting k8s for", id, ". assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	// Ingress
	for mapName, ingress := range vm.Subsystems.K8s.IngressMap {
		err = resources.SsDeleter(kc.DeleteIngress).
			WithResourceID(ingress.Name).
			WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for mapName, k8sService := range vm.Subsystems.K8s.ServiceMap {
		err = resources.SsDeleter(kc.DeleteService).
			WithResourceID(k8sService.Name).
			WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for mapName, k8sDeployment := range vm.Subsystems.K8s.DeploymentMap {
		err = resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.Name).
			WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for mapName, secret := range vm.Subsystems.K8s.SecretMap {
		var deleteFunc func(id string) error
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
		WithResourceID(vm.Subsystems.K8s.Namespace.Name).
		WithDbFunc(dbFunc(id, "namespace")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

// Repair repairs the K8s setup for a VM.
func (c *Client) Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair k8s %s. details: %w", id, err)
	}

	vm, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			log.Println("vm not found when deleting k8s for", id, ". assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	namespace := g.Namespace()
	if namespace != nil {
		err = resources.SsRepairer(
			kc.ReadNamespace,
			kc.CreateNamespace,
			kc.UpdateNamespace,
			func(string) error { return nil },
		).WithResourceID(namespace.Name).WithDbFunc(dbFunc(id, "namespace")).WithGenPublic(namespace).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	deployments := g.Deployments()
	for mapName, k8sDeployment := range vm.Subsystems.K8s.DeploymentMap {
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
	for mapName, k8sService := range vm.Subsystems.K8s.ServiceMap {
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
	for mapName, ingress := range vm.Subsystems.K8s.IngressMap {
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
	for mapName, secret := range vm.Subsystems.K8s.SecretMap {
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

// EnsureOwner ensures the owner of the K8s setup, by deleting and then trigger a call to Repair.
func (c *Client) EnsureOwner(id, oldOwnerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s owner for vm %s. details: %w", id, err)
	}

	// Since ownership is determined by the namespace, and the namespace owns everything,
	// we need to recreate everything

	// Delete everything in the old namespace
	// Pass in the old owner ID to use the old namespace
	err := c.Delete(id)
	if err != nil {
		return makeError(err)
	}

	// Create everything in the new namespace
	// We reset the namespace to use the VM's namespace by passing an empty string
	err = c.Repair(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// dbFunc returns a function that updates the K8s subsystem.
func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return vmModels.New(versions.V1).DeleteSubsystem(id, "k8s."+key)
		}
		return vmModels.New(versions.V1).SetSubsystem(id, "k8s."+key, data)
	}
}
