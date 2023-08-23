package k8s_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/utils/subsystemutils"
	"log"
	"path"
	"reflect"
	"strconv"
)

type K8sResult struct {
	Namespace  *k8sModels.NamespacePublic
	Deployment *k8sModels.DeploymentPublic
	Service    *k8sModels.ServicePublic
	Ingress    *k8sModels.IngressPublic
}

func Create(deploymentID string, userID string, params *deploymentModel.CreateParams) (*K8sResult, error) {
	log.Println("setting up k8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %s", params.Name, err)
	}

	deployment, err := deploymentModel.GetByID(deploymentID)
	if err != nil {
		return nil, makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", deploymentID, "not found for k8s setup assuming it was deleted")
		return nil, nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return nil, fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return nil, makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	// Namespace
	namespace := &ss.Namespace
	if !ss.Namespace.Created() {
		public := createNamespacePublic(userID)
		namespace, err = createNamespace(client, deployment.ID, ss, public, deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return nil, makeError(err)
		}
	}

	// PersistentVolume
	if ss.PvMap == nil {
		ss.PvMap = make(map[string]k8sModels.PvPublic)
	}

	for _, volume := range params.Volumes {
		k8sName := fmt.Sprintf("%s-%s", deployment.Name, volume.Name)
		pv, exists := ss.PvMap[volume.Name]
		if !pv.Created() || !exists {
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			nfsPath := path.Join(zone.Storage.NfsParentPath, deployment.OwnerID, "user", volume.ServerPath)

			public := createPvPublic(k8sName, capacity, nfsPath, zone.Storage.NfsServer)
			_, err = createPV(client, deployment.ID, volume.Name, ss, public, deploymentModel.UpdateSubsystemByID)
			if err != nil {
				return nil, makeError(err)
			}
		}
	}

	// PersistentVolumeClaim
	if ss.PvcMap == nil {
		ss.PvcMap = make(map[string]k8sModels.PvcPublic)
	}

	for _, volume := range params.Volumes {
		k8sName := fmt.Sprintf("%s-%s", deployment.Name, volume.Name)
		pvc, exists := ss.PvcMap[volume.Name]
		if !pvc.Created() || !exists {
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			public := createPvcPublic(namespace.FullName, k8sName, capacity, k8sName)
			_, err = createPVC(client, deployment.ID, volume.Name, ss, public, deploymentModel.UpdateSubsystemByID)
			if err != nil {
				return nil, makeError(err)
			}
		}
	}

	// Deployment
	k8sDeployment := &ss.Deployment
	if !ss.Deployment.Created() {
		dockerRegistryProject := subsystemutils.GetPrefixedName(userID)
		dockerImage := fmt.Sprintf("%s/%s/%s", conf.Env.DockerRegistry.URL, dockerRegistryProject, deployment.Name)
		public := createDeploymentPublic(namespace.FullName, deployment.Name, dockerImage, params.Envs, params.Volumes, params.InitCommands)
		k8sDeployment, err = createK8sDeployment(client, deployment.ID, ss, public, deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return nil, makeError(err)
		}
	}

	// Service
	service := &ss.Service
	if !ss.Service.Created() {
		public := createServicePublic(namespace.FullName, deployment.Name, conf.Env.Deployment.Port)
		service, err = createService(client, deployment.ID, ss, public, deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return nil, makeError(err)
		}
	}

	// Ingress
	ingress := &ss.Ingress
	if params.Private {
		if ss.Ingress.Created() {
			err = client.DeleteIngress(ss.Ingress.Namespace, ss.Ingress.Name)
			if err != nil {
				return nil, makeError(err)
			}
		}
		ingress = &k8sModels.IngressPublic{
			Placeholder: true,
		}

		err = deploymentModel.UpdateSubsystemByName(deployment.Name, "k8s", "ingress", ingress)
		if err != nil {
			return nil, makeError(err)
		}

	} else if !ss.Ingress.Created() {
		ingress, err = createIngress(client, deployment.ID, ss, createIngressPublic(
			namespace.FullName,
			deployment.Name,
			service.Name,
			service.Port,
			[]string{getExternalFQDN(deployment.Name, zone)},
		), deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return nil, makeError(err)
		}
	}

	return &K8sResult{
		Namespace:  namespace,
		Deployment: k8sDeployment,
		Service:    service,
		Ingress:    ingress,
	}, nil
}

func Delete(name string) error {
	log.Println("deleting k8s for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %s", name, err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found for k8s deletion. assuming it was deleted")
		return nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	// Ingress
	if ss.Ingress.Created() {
		err = deleteIngress(client, deployment.ID, ss, deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	if ss.Service.Created() {
		err = deleteService(client, deployment.ID, ss, deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	if ss.Deployment.Created() {
		err = deleteDeployment(client, deployment.ID, ss, deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	if ss.PvcMap != nil {
		for pvcName, pvc := range ss.PvcMap {
			if pvc.Created() {
				err = deletePVC(client, deployment.ID, pvcName, ss, deploymentModel.UpdateSubsystemByID)
				if err != nil {
					return makeError(err)
				}
			}
		}
	}

	// PersistentVolume
	if ss.PvMap != nil {
		for pvName, pv := range ss.PvMap {
			if pv.Created() {
				err = deletePV(client, deployment.ID, pvName, ss, deploymentModel.UpdateSubsystemByID)
				if err != nil {
					return makeError(err)
				}
			}
		}
	}

	// Job
	if ss.JobMap != nil {
		for jobName, job := range ss.JobMap {
			if job.Created() {
				err = deleteJob(client, deployment.ID, jobName, ss, deploymentModel.UpdateSubsystemByID)
				if err != nil {
					return makeError(err)
				}
			}
		}
	}

	return nil
}

func Update(name string, params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s for deployment %s. details: %s", name, err)
	}

	if params == nil || (params.Envs == nil && params.Private == nil) {
		return nil
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found for k8s update assuming it was deleted")
		return nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems
	updateDb := func(id, subsystem, key string, update interface{}) error {
		return deploymentModel.UpdateSubsystemByID(id, subsystem, key, update)
	}

	if params.Envs != nil {
		if ss.K8s.Deployment.Created() {
			k8sEnvs := []k8sModels.EnvVar{
				{Name: "DEPLOY_APP_PORT", Value: strconv.Itoa(conf.Env.Deployment.Port)},
			}
			for _, env := range *params.Envs {
				k8sEnvs = append(k8sEnvs, k8sModels.EnvVar{
					Name:  env.Name,
					Value: env.Value,
				})
			}

			ss.K8s.Deployment.EnvVars = k8sEnvs

			err = client.UpdateDeployment(&ss.K8s.Deployment)
			if err != nil {
				return makeError(err)
			}

			err = deploymentModel.UpdateSubsystemByName(name, "k8s", "deployment", &ss.K8s.Deployment)
			if err != nil {
				return makeError(err)
			}
		}
	}

	if params.ExtraDomains != nil {
		if ss.K8s.Ingress.Created() {
			ingress := ss.K8s.Ingress
			if ingress.ID == "" {
				return nil
			}

			newPublic := &ss.K8s.Ingress
			newPublic.Hosts = *params.ExtraDomains

			err = recreateIngress(client, deployment.ID, &ss.K8s, newPublic, updateDb)
			if err != nil {
				return makeError(err)
			}
		}
	}

	if params.Private != nil {
		emptyOrPlaceHolder := !ss.K8s.Ingress.Created() || ss.K8s.Ingress.Placeholder

		if *params.Private != emptyOrPlaceHolder {
			if !emptyOrPlaceHolder {
				err = client.DeleteIngress(ss.K8s.Ingress.Namespace, ss.K8s.Ingress.ID)
				if err != nil {
					return makeError(err)
				}

				err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{})
				if err != nil {
					return makeError(err)
				}

				deployment, err = deploymentModel.GetByName(name)
				if err != nil {
					return makeError(err)
				}
			}

			if *params.Private {
				err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{
					Placeholder: true,
				})
				if err != nil {
					return makeError(err)
				}
			} else {
				namespace := ss.K8s.Namespace
				if !namespace.Created() {
					return nil
				}

				service := ss.K8s.Service
				if !service.Created() {
					return nil
				}

				var domains []string
				if params.ExtraDomains == nil {
					domains = getAllDomainNames(deployment.Name, deployment.ExtraDomains, zone)
				} else {
					domains = getAllDomainNames(deployment.Name, *params.ExtraDomains, zone)
				}

				public := createIngressPublic(namespace.FullName, name, service.Name, service.Port, domains)
				_, err = createIngress(client, deployment.ID, &ss.K8s, public, updateDb)
				if err != nil {
					return makeError(err)
				}

			}
		}
	}
	return nil
}

func Restart(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s %s. details: %s", name, err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if !deployment.Subsystems.K8s.Deployment.Created() {
		return makeError(errors.New("can't restart deployment that is not yet created"))
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	err = client.RestartDeployment(&deployment.Subsystems.K8s.Deployment)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Repair(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair k8s %s. details: %s", name, err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found when repairing k8s, assuming it was deleted")
		return nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	// temporary fix for missing resource limits and requests
	if ss.Deployment.Created() {
		res := &ss.Deployment.Resources
		if res.Limits.Memory == "" && res.Limits.CPU == "" && res.Requests.Memory == "" && res.Requests.CPU == "" {
			res.Limits.Memory = conf.Env.Deployment.Resources.Limits.Memory
			res.Limits.CPU = conf.Env.Deployment.Resources.Limits.CPU
			res.Requests.Memory = conf.Env.Deployment.Resources.Requests.Memory
			res.Requests.CPU = conf.Env.Deployment.Resources.Requests.CPU
		}
	}

	// namespace
	namespace, err := client.ReadNamespace(ss.Namespace.ID)
	if err != nil {
		return makeError(err)
	}

	if namespace == nil || !reflect.DeepEqual(ss.Namespace, *namespace) {
		log.Println("recreating namespace for deployment", name)
		err = recreateNamespace(client, deployment.ID, ss, &ss.Namespace, deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// deployment
	k8sDeployment, err := client.ReadDeployment(ss.Namespace.FullName, ss.Deployment.ID)
	if err != nil {
		return makeError(err)
	}

	if k8sDeployment == nil || !reflect.DeepEqual(ss.Deployment, *k8sDeployment) {
		log.Println("recreating deployment for deployment", name)
		err = recreateK8sDeployment(client, deployment.ID, ss, &ss.Deployment, deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// service
	service, err := client.ReadService(ss.Namespace.FullName, ss.Service.ID)
	if err != nil {
		return makeError(err)
	}

	if service == nil || !reflect.DeepEqual(ss.Service, *service) {
		log.Println("recreating service for deployment", name)
		err = recreateService(client, deployment.ID, ss, &ss.Service, deploymentModel.UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// ingress
	if deployment.Private != ss.Ingress.Placeholder {
		log.Println("recreating ingress for deployment due to mismatch with the private field", name)

		if deployment.Private {
			err = client.DeleteIngress(deployment.Subsystems.K8s.Ingress.Namespace, deployment.Subsystems.K8s.Ingress.ID)
			if err != nil {
				return makeError(err)
			}

			err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{
				Placeholder: true,
			})
			if err != nil {
				return makeError(err)
			}
		} else {
			_, err = createIngress(client, deployment.ID, ss, createIngressPublic(
				deployment.Subsystems.K8s.Namespace.FullName,
				deployment.Name,
				deployment.Subsystems.K8s.Service.Name,
				deployment.Subsystems.K8s.Service.Port,
				getAllDomainNames(deployment.Name, deployment.ExtraDomains, zone),
			), deploymentModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	} else if !ss.Ingress.Placeholder {
		ingress, err := client.ReadIngress(ss.Namespace.FullName, ss.Ingress.ID)
		if err != nil {
			return makeError(err)
		}

		if ingress == nil || !reflect.DeepEqual(ss.Ingress, *ingress) {
			log.Println("recreating ingress for deployment", name)
			err = recreateIngress(client, deployment.ID, ss, &ss.Ingress, deploymentModel.UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	return nil
}
