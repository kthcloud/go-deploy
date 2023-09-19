package k8s_service

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/utils/subsystemutils"
	"log"
	"path"
	"strconv"
)

const (
	appName = "main"
)

func withClient(zone *enviroment.DeploymentZone, namespace string) (*k8s.Client, error) {
	client, err := k8s.New(zone.Client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	if namespace != "" {
		namespaceCreated, err := client.NamespaceCreated(namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to check if namespace %s is created. details: %w", namespace, err)
		}

		if !namespaceCreated {
			return nil, fmt.Errorf("no such namespace %s", namespace)
		}
	}

	return client, nil
}

func Create(deploymentID string, userID string, params *deploymentModel.CreateParams) error {
	log.Println("setting up k8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %w", params.Name, err)
	}

	deployment, err := deploymentModel.New().GetByID(deploymentID)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", deploymentID, "not found for k8s setup assuming it was deleted")
		return nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return fmt.Errorf("main app not found for deployment %s", deploymentID)
	}

	client, err := withClient(zone, subsystemutils.GetPrefixedName(userID))
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	// Namespace
	if service.NotCreated(&ss.Namespace) {
		public := createNamespacePublic(userID)
		_, err = createNamespace(client, deployment.ID, ss, public, deploymentModel.New().UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	for _, volume := range params.Volumes {
		if service.NotCreated(ss.GetPV(volume.Name)) {
			k8sName := fmt.Sprintf("%s-%s", deployment.Name, volume.Name)
			nfsPath := path.Join(zone.Storage.NfsParentPath, deployment.OwnerID, "user", volume.ServerPath)

			public := createPvPublic(k8sName, conf.Env.Deployment.Resources.Limits.Storage, nfsPath, zone.Storage.NfsServer)
			_, err = createPV(client, deployment.ID, volume.Name, ss, public, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolumeClaim
	for _, volume := range params.Volumes {
		if service.NotCreated(ss.GetPVC(volume.Name)) {
			k8sName := fmt.Sprintf("%s-%s", deployment.Name, volume.Name)

			public := createPvcPublic(client.Namespace, k8sName, conf.Env.Deployment.Resources.Limits.Storage, k8sName)
			_, err = createPVC(client, deployment.ID, volume.Name, ss, public, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Deployment
	if service.NotCreated(ss.GetDeployment(appName)) {
		public := createMainAppDeploymentPublic(client.Namespace, deployment.Name, deployment.OwnerID, mainApp.InternalPort, params.Envs, params.Volumes, params.InitCommands)
		_, err = createK8sDeployment(client, deployment.ID, appName, ss, public, deploymentModel.New().UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	if service.NotCreated(ss.GetService(appName)) {
		public := createServicePublic(client.Namespace, deployment.Name, conf.Env.Deployment.Port, mainApp.InternalPort)
		_, err = createService(client, deployment.ID, appName, ss, public, deploymentModel.New().UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	if ingress := ss.GetIngress(appName); service.NotCreated(ingress) {

		var public *k8sModels.IngressPublic
		if params.Private {
			public = &k8sModels.IngressPublic{
				Placeholder: true,
			}
		} else {
			public = createIngressPublic(
				client.Namespace,
				deployment.Name,
				ss.GetService(appName).Name,
				ss.GetService(appName).Port,
				[]string{getExternalFQDN(deployment.Name, zone)},
			)
		}

		_, err = createIngress(client, deployment.ID, appName, ss, public, deploymentModel.New().UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func Delete(name string) error {
	log.Println("deleting k8s for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %w", name, err)
	}

	deployment, err := deploymentModel.New().GetByName(name)
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

	client, err := k8s.New(zone.Client, subsystemutils.GetPrefixedName(deployment.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	// Ingress
	for mapName, ingress := range ss.IngressMap {
		if ingress.Created() {
			err = deleteIngress(client, deployment.ID, mapName, ss, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Service
	for mapName, k8sService := range ss.ServiceMap {
		if k8sService.Created() {
			err = deleteService(client, deployment.ID, mapName, ss, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Deployment
	for mapName, k8sDeployment := range ss.DeploymentMap {
		if k8sDeployment.Created() {
			err = deleteK8sDeployment(client, deployment.ID, mapName, ss, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolumeClaim
	for pvcName, pvc := range ss.PvcMap {
		if pvc.Created() {
			err = deletePVC(client, deployment.ID, pvcName, ss, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolume
	for mapName, pv := range ss.PvMap {
		if pv.Created() {
			err = deletePV(client, deployment.ID, mapName, ss, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Job
	for mapName, job := range ss.JobMap {
		if job.Created() {
			err = deleteJob(client, deployment.ID, mapName, ss, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Namespace
	if ss.Namespace.Created() {
		err = deleteNamespace(client, deployment.ID, ss, deploymentModel.New().UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func Update(name string, params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s for deployment %s. details: %w", name, err)
	}

	if *params == (deploymentModel.UpdateParams{}) {
		return nil
	}

	deployment, err := deploymentModel.New().GetByName(name)
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

	client, err := k8s.New(zone.Client, subsystemutils.GetPrefixedName(deployment.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems
	updateDb := func(id, subsystem, key string, update interface{}) error {
		return deploymentModel.New().UpdateSubsystemByID(id, subsystem, key, update)
	}

	mainApp := deployment.GetMainApp()

	if params.Envs != nil {
		k8sDeployment, ok := ss.K8s.DeploymentMap[appName]
		if ok && k8sDeployment.Created() {
			k8sEnvs := []k8sModels.EnvVar{
				{Name: "PORT", Value: strconv.Itoa(mainApp.InternalPort)},
			}
			for _, env := range *params.Envs {
				k8sEnvs = append(k8sEnvs, k8sModels.EnvVar{
					Name:  env.Name,
					Value: env.Value,
				})
			}

			k8sDeployment.EnvVars = k8sEnvs

			err = client.UpdateDeployment(&k8sDeployment)
			if err != nil {
				return makeError(err)
			}

			ss.K8s.DeploymentMap[appName] = k8sDeployment

			err = deploymentModel.New().UpdateSubsystemByName(name, "k8s", "deploymentMap", &ss.K8s.DeploymentMap)
			if err != nil {
				return makeError(err)
			}
		}
	}

	if params.ExtraDomains != nil {
		ingress, ok := ss.K8s.IngressMap[appName]
		if ok && ingress.Created() {
			if ingress.ID == "" {
				return nil
			}

			ingress.Hosts = *params.ExtraDomains

			err = recreateIngress(client, deployment.ID, name, &ss.K8s, &ingress, updateDb)
			if err != nil {
				return makeError(err)
			}
		}
	}

	if params.Private != nil {
		ingress := ss.K8s.IngressMap[appName]

		emptyOrPlaceHolder := !ingress.Created() || ingress.IsPlaceholder()

		if *params.Private != emptyOrPlaceHolder {
			if !emptyOrPlaceHolder {
				err = deleteIngress(client, deployment.ID, appName, &ss.K8s, updateDb)
				if err != nil {
					return makeError(err)
				}
			}

			if *params.Private {
				ss.K8s.IngressMap[appName] = k8sModels.IngressPublic{
					Placeholder: true,
				}

				err = deploymentModel.New().UpdateSubsystemByName(name, "k8s", "ingressMap", ss.K8s.IngressMap)
				if err != nil {
					return makeError(err)
				}
			} else {
				namespace := ss.K8s.Namespace
				if !namespace.Created() {
					return nil
				}

				k8sService := ss.K8s.GetService(appName)
				if service.NotCreated(k8sService) {
					return nil
				}

				var domains []string
				if params.ExtraDomains == nil {
					domains = getAllDomainNames(deployment.Name, mainApp.ExtraDomains, zone)
				} else {
					domains = getAllDomainNames(deployment.Name, *params.ExtraDomains, zone)
				}

				public := createIngressPublic(namespace.FullName, name, k8sService.Name, k8sService.Port, domains)
				_, err = createIngress(client, deployment.ID, appName, &ss.K8s, public, updateDb)
				if err != nil {
					return makeError(err)
				}

			}
		}
	}

	if params.Volumes != nil {
		// delete deployment, pvcs and pvs
		// then
		// create new deployment, pvcs and pvs

		k8sDeployment, ok := ss.K8s.DeploymentMap[appName]
		if ok && k8sDeployment.Created() {
			err = deleteK8sDeployment(client, deployment.ID, appName, &ss.K8s, updateDb)
			if err != nil {
				return makeError(err)
			}
		}

		for mapName, pvc := range ss.K8s.PvcMap {
			if pvc.Created() {
				err = deletePVC(client, pvc.ID, mapName, &ss.K8s, updateDb)
				if err != nil {
					return makeError(err)
				}
			}
		}

		for mapName, pv := range ss.K8s.PvMap {
			if pv.Created() {
				err = deletePV(client, pv.ID, mapName, &ss.K8s, updateDb)
				if err != nil {
					return makeError(err)
				}
			}
		}

		// clear the maps
		ss.K8s.DeploymentMap = make(map[string]k8sModels.DeploymentPublic)
		ss.K8s.PvcMap = make(map[string]k8sModels.PvcPublic)
		ss.K8s.PvMap = make(map[string]k8sModels.PvPublic)

		// since we depend on the namespace, we must ensure it is actually created here
		if !ss.K8s.Namespace.Created() {
			public := createNamespacePublic(deployment.OwnerID)
			namespace, err := createNamespace(client, deployment.ID, &ss.K8s, public, updateDb)
			if err != nil {
				return makeError(err)
			}
			ss.K8s.Namespace = *namespace
		}

		for _, volume := range *params.Volumes {
			k8sName := fmt.Sprintf("%s-%s", deployment.Name, volume.Name)
			capacity := conf.Env.Deployment.Resources.Limits.Storage
			nfsPath := path.Join(zone.Storage.NfsParentPath, deployment.OwnerID, "user", volume.ServerPath)

			pvPublic := createPvPublic(k8sName, capacity, nfsPath, zone.Storage.NfsServer)
			_, err = createPV(client, deployment.ID, volume.Name, &ss.K8s, pvPublic, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}

			pvcPublic := createPvcPublic(ss.K8s.Namespace.FullName, k8sName, capacity, k8sName)
			_, err = createPVC(client, deployment.ID, volume.Name, &ss.K8s, pvcPublic, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}

		public := createMainAppDeploymentPublic(ss.K8s.Namespace.FullName, deployment.Name, deployment.OwnerID, mainApp.InternalPort, mainApp.Envs, *params.Volumes, mainApp.InitCommands)
		_, err = createK8sDeployment(client, deployment.ID, appName, &ss.K8s, public, deploymentModel.New().UpdateSubsystemByID)
		if err != nil {
			return makeError(err)
		}

	}
	return nil
}

func Restart(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s %s. details: %w", name, err)
	}

	deployment, err := deploymentModel.New().GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found for k8s restart. assuming it was deleted")
		return nil
	}

	k8sDeployment, ok := deployment.Subsystems.K8s.DeploymentMap["main"]
	if !ok || !k8sDeployment.Created() {
		log.Println("k8s deployment not created when restarting. assuming it was deleted")
		return nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client, subsystemutils.GetPrefixedName(deployment.OwnerID))
	if err != nil {
		return makeError(err)
	}

	err = client.RestartDeployment(&k8sDeployment)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Repair(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair k8s %s. details: %w", name, err)
	}

	deployment, err := deploymentModel.New().GetByName(name)
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

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		log.Println("main app not created when repairing k8s assuming it was deleted")
		return nil
	}

	client, err := k8s.New(zone.Client, subsystemutils.GetPrefixedName(deployment.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	err = repairNamespace(client, deployment.ID, ss, deploymentModel.New().UpdateSubsystemByID, func() *k8sModels.NamespacePublic {
		return createNamespacePublic(deployment.OwnerID)
	})

	if err != nil {
		return makeError(err)
	}

	for mapName := range ss.DeploymentMap {
		err = repairDeployment(client, deployment.ID, mapName, ss, deploymentModel.New().UpdateSubsystemByID, func() *k8sModels.DeploymentPublic {
			if mapName == appName {
				return createMainAppDeploymentPublic(
					ss.Namespace.FullName,
					deployment.Name,
					deployment.OwnerID,
					mainApp.InternalPort,
					mainApp.Envs,
					mainApp.Volumes,
					mainApp.InitCommands,
				)
			}
			return nil
		})
		if err != nil {
			return makeError(err)
		}
	}

	for mapName := range ss.ServiceMap {
		err = repairService(client, deployment.ID, mapName, ss, deploymentModel.New().UpdateSubsystemByID, func() *k8sModels.ServicePublic {
			if mapName == appName {
				return createServicePublic(
					ss.Namespace.FullName,
					deployment.Name,
					conf.Env.Deployment.Port,
					mainApp.InternalPort,
				)
			}
			return nil
		})
		if err != nil {
			return makeError(err)
		}
	}

	mainIngress := ss.GetIngress(appName)
	if service.NotCreated(mainIngress) {
		log.Println("main ingress not created when recreating ingress. assuming it was deleted")
		return nil
	}

	if mainApp.Private != mainIngress.Placeholder {
		log.Println("recreating ingress for deployment due to mismatch with the private field", name)

		if mainApp.Private {
			err = deleteIngress(client, deployment.ID, appName, ss, deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		} else {
			k8sService := ss.GetService(appName)
			if service.NotCreated(k8sService) {
				log.Println("main service not created when recreating ingress. assuming it was deleted")
				return nil
			}

			_, err = createIngress(client, deployment.ID, appName, ss, createIngressPublic(
				deployment.Subsystems.K8s.Namespace.FullName,
				deployment.Name,
				k8sService.Name,
				k8sService.Port,
				getAllDomainNames(deployment.Name, mainApp.ExtraDomains, zone),
			), deploymentModel.New().UpdateSubsystemByID)
			if err != nil {
				return makeError(err)
			}
		}
	} else if !mainIngress.Placeholder {
		err = repairIngress(client, deployment.ID, appName, ss, deploymentModel.New().UpdateSubsystemByID, func() *k8sModels.IngressPublic {
			return createIngressPublic(
				deployment.Subsystems.K8s.Namespace.FullName,
				deployment.Name,
				mainIngress.ServiceName,
				mainIngress.ServicePort,
				mainIngress.Hosts,
			)
		})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}
