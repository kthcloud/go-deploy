package k8s_service

import (
	"fmt"
	"go-deploy/models/sys/storage_manager"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	"go-deploy/service/resources"
	"go-deploy/service/storage_manager_service/client"
	"golang.org/x/exp/slices"
	"log"
)

// Create creates the storage manager.
//
// It creates all K8s resources for the storage manager.
func (c *Client) Create(params *storage_manager.CreateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create storage manager in k8s. details: %w", err)
	}

	_, kc, g, err := c.Get(client.OptsNoStorageManager)
	if err != nil {
		return makeError(err)
	}

	// Namespace
	err = resources.SsCreator(kc.CreateNamespace).
		WithDbFunc(dbFunc(c.ID(), "namespace")).
		WithPublic(g.Namespace()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// PersistentVolume
	for _, pvPublic := range g.PVs() {
		err = resources.SsCreator(kc.CreatePV).
			WithDbFunc(dbFunc(c.ID(), "pvMap."+pvPublic.Name)).
			WithPublic(&pvPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for _, pvcPublic := range g.PVCs() {
		err = resources.SsCreator(kc.CreatePVC).
			WithDbFunc(dbFunc(c.ID(), "pvcMap."+pvcPublic.Name)).
			WithPublic(&pvcPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Job
	for _, jobPublic := range g.Jobs() {
		err = resources.SsCreator(kc.CreateJob).
			WithDbFunc(dbFunc(c.ID(), "jobMap."+jobPublic.Name)).
			WithPublic(&jobPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for _, secret := range g.Secrets() {
		err = resources.SsCreator(kc.CreateSecret).
			WithDbFunc(dbFunc(c.ID(), "secretMap."+secret.Name)).
			WithPublic(&secret).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deployment := range g.Deployments() {
		err = resources.SsCreator(kc.CreateDeployment).
			WithDbFunc(dbFunc(c.ID(), "deploymentMap."+deployment.Name)).
			WithPublic(&deployment).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, k8sService := range g.Services() {
		err = resources.SsCreator(kc.CreateService).
			WithDbFunc(dbFunc(c.ID(), "serviceMap."+k8sService.Name)).
			WithPublic(&k8sService).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	for _, ingress := range g.Ingresses() {
		err = resources.SsCreator(kc.CreateIngress).
			WithDbFunc(dbFunc(c.ID(), "ingressMap."+ingress.Name)).
			WithPublic(&ingress).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

// Delete deletes the storage manager.
//
// It deletes all K8s resources for the storage manager.
func (c *Client) Delete() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete storage manager in k8s. details: %w", err)
	}

	log.Println("deleting k8s for storage manager", c.ID())

	sm, kc, _, err := c.Get(client.OptsNoGenerator)
	if err != nil {
		return makeError(err)
	}

	// Deployment
	for mapName, k8sDeployment := range sm.Subsystems.K8s.DeploymentMap {
		err = resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.ID).
			WithDbFunc(dbFunc(c.ID(), "deploymentMap."+mapName)).
			Exec()
	}

	// Service
	for mapName, k8sService := range sm.Subsystems.K8s.ServiceMap {
		err = resources.SsDeleter(kc.DeleteService).
			WithResourceID(k8sService.ID).
			WithDbFunc(dbFunc(c.ID(), "serviceMap."+mapName)).
			Exec()
	}

	// Ingress
	for mapName, ingress := range sm.Subsystems.K8s.IngressMap {
		err = resources.SsDeleter(kc.DeleteIngress).
			WithResourceID(ingress.ID).
			WithDbFunc(dbFunc(c.ID(), "ingressMap."+mapName)).
			Exec()
	}

	// Job
	for mapName, job := range sm.Subsystems.K8s.JobMap {
		err = resources.SsDeleter(kc.DeleteJob).
			WithResourceID(job.ID).
			WithDbFunc(dbFunc(c.ID(), "jobMap."+mapName)).
			Exec()
	}

	// PersistentVolumeClaim
	for mapName, pvc := range sm.Subsystems.K8s.PvcMap {
		err = resources.SsDeleter(kc.DeletePVC).
			WithResourceID(pvc.ID).
			WithDbFunc(dbFunc(c.ID(), "pvcMap."+mapName)).
			Exec()
	}

	// PersistentVolume
	for mapName, pv := range sm.Subsystems.K8s.PvMap {
		err = resources.SsDeleter(kc.DeletePV).
			WithResourceID(pv.ID).
			WithDbFunc(dbFunc(c.ID(), "pvMap."+mapName)).
			Exec()
	}

	// Secret
	for mapName, secret := range sm.Subsystems.K8s.SecretMap {
		var deleteFunc func(interface{}) error
		if mapName == constants.WildcardCertSecretName {
			deleteFunc = func(interface{}) error { return nil }
		} else {
			deleteFunc = dbFunc(c.ID(), "secretMap."+mapName)
		}

		err = resources.SsDeleter(kc.DeleteSecret).
			WithResourceID(secret.ID).
			WithDbFunc(deleteFunc).
			Exec()
	}

	return nil
}

// Repair repairs the storage manager.
//
// It repairs all K8s resources for the storage manager.
func (c *Client) Repair() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair storage manager %s in k8s. details: %w", c.ID(), err)
	}

	sm, kc, g, err := c.Get(client.OptsAll)
	if err != nil {
		return makeError(err)
	}

	namespace := g.Namespace()
	err = resources.SsRepairer(
		kc.ReadNamespace,
		kc.CreateNamespace,
		kc.UpdateNamespace,
		func(string) error { return nil },
	).WithResourceID(namespace.ID).WithDbFunc(dbFunc(c.ID(), "namespace")).WithGenPublic(namespace).Exec()

	if err != nil {
		return makeError(err)
	}

	deployments := g.Deployments()
	for mapName, k8sDeployment := range sm.Subsystems.K8s.DeploymentMap {
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
	for mapName, k8sService := range sm.Subsystems.K8s.ServiceMap {
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
	for mapName, ingress := range sm.Subsystems.K8s.IngressMap {
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
	for mapName, secret := range sm.Subsystems.K8s.SecretMap {
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

	jobs := g.Jobs()
	for mapName, job := range sm.Subsystems.K8s.JobMap {
		idx := slices.IndexFunc(jobs, func(j k8sModels.JobPublic) bool { return j.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteJob).
				WithResourceID(job.ID).
				WithDbFunc(dbFunc(c.ID(), "jobMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range jobs {
		err = resources.SsRepairer(
			kc.ReadJob,
			kc.CreateJob,
			func(job *k8sModels.JobPublic) (*k8sModels.JobPublic, error) { return nil, nil },
			kc.DeleteJob,
		).WithResourceID(public.ID).WithDbFunc(dbFunc(c.ID(), "jobMap."+public.Name)).WithGenPublic(&public).Exec()

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
			return storage_manager.New().DeleteSubsystemByID(id, "k8s."+key)
		}
		return storage_manager.New().UpdateSubsystemByID(id, "k8s."+key, data)
	}
}
