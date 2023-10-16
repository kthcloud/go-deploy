package k8s_service

import (
	"errors"
	"fmt"
	vmModels "go-deploy/models/sys/vm"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/base"
	"golang.org/x/exp/slices"
	"log"
)

func Create(id string, params *vmModels.CreateParams) error {
	log.Println("setting up k8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %w", params.Name, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.VmDeletedErr) {
			log.Println("vm not found when creating k8s for deployment", id, ". assuming it was deleted")
			return nil
		}

		if errors.Is(err, base.DeploymentZoneNotFoundErr) {
			log.Println("deployment zone not found for deployment")
			return nil
		}

		return makeError(err)
	}

	context.WithCreateParams(params)

	// Namespace
	err = resources.SsCreator(context.Client.CreateNamespace).
		WithDbFunc(dbFunc(id, "namespace")).
		WithPublic(context.Generator.Namespace()).
		Exec()
	if err != nil {
		return makeError(err)
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
	for mapName, ingress := range context.VM.Subsystems.K8s.IngressMap {
		err = resources.SsDeleter(context.Client.DeleteIngress).
			WithResourceID(ingress.ID).
			WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
			Exec()
	}

	// Service
	for mapName, k8sService := range context.VM.Subsystems.K8s.ServiceMap {
		err = resources.SsDeleter(context.Client.DeleteService).
			WithResourceID(k8sService.ID).
			WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
			Exec()
	}

	// Deployment
	for mapName, k8sDeployment := range context.VM.Subsystems.K8s.DeploymentMap {
		err = resources.SsDeleter(context.Client.DeleteDeployment).
			WithResourceID(k8sDeployment.ID).
			WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
			Exec()
	}

	// Namespace
	err = resources.SsDeleter(func(string) error { return nil }).
		WithResourceID(context.VM.Subsystems.K8s.Namespace.ID).
		WithDbFunc(dbFunc(id, "namespace")).
		Exec()

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
	for mapName, k8sDeployment := range context.VM.Subsystems.K8s.DeploymentMap {
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
	for mapName, k8sService := range context.VM.Subsystems.K8s.ServiceMap {
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
	for mapName, ingress := range context.VM.Subsystems.K8s.IngressMap {
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

	return nil
}

func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return vmModels.New().DeleteSubsystemByID(id, "k8s."+key)
		}
		return vmModels.New().UpdateSubsystemByID(id, "k8s."+key, data)
	}
}
