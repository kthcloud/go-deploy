package k8s_service

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/deployment_service/k8s_service/helpers"
	"go-deploy/utils/subsystemutils"
	"log"
	"path"
	"strconv"
	"strings"
)

const (
	appName                = "main"
	appNameCustomDomain    = "custom-domain"
	appNameImagePullSecret = "image-pull-secret"
)

var CustomDomainInUseErr = fmt.Errorf("custom domain is already in use")

func withCustomDomainSuffix(name string) string {
	return fmt.Sprintf("%s-%s", name, appNameCustomDomain)
}

func withImagePullSecretSuffix(name string) string {
	return fmt.Sprintf("%s-%s", name, appNameImagePullSecret)
}

func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

func Create(deploymentID string, params *deploymentModel.CreateParams) error {
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

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return fmt.Errorf("main app not found for deployment %s", deploymentID)
	}

	client, err := helpers.New(&deployment.Subsystems.K8s, deployment.Zone, getNamespaceName(deployment.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	// Namespace
	if service.NotCreated(&ss.Namespace) {
		public := helpers.CreateNamespacePublic(deployment.OwnerID)
		_, err = client.CreateNamespace(deployment.ID, public)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	for _, volume := range params.Volumes {
		if service.NotCreated(ss.GetPV(volume.Name)) {
			k8sName := fmt.Sprintf("%s-%s", deployment.Name, volume.Name)
			nfsPath := path.Join(client.Zone.Storage.NfsParentPath, deployment.OwnerID, "user", volume.ServerPath)

			public := helpers.CreatePvPublic(k8sName, conf.Env.Deployment.Resources.Limits.Storage, nfsPath, client.Zone.Storage.NfsServer)
			_, err = client.CreatePV(deployment.ID, volume.Name, public)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// PersistentVolumeClaim
	for _, volume := range params.Volumes {
		if service.NotCreated(ss.GetPVC(volume.Name)) {
			k8sName := fmt.Sprintf("%s-%s", deployment.Name, volume.Name)

			public := helpers.CreatePvcPublic(client.Namespace, k8sName, conf.Env.Deployment.Resources.Limits.Storage, k8sName)
			_, err = client.CreatePVC(deployment.ID, volume.Name, public)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// Secret
	secret := ss.GetSecret(appName)
	if service.NotCreated(secret) {
		registry := conf.Env.DockerRegistry.URL
		username := deployment.Subsystems.Harbor.Robot.HarborName
		password := deployment.Subsystems.Harbor.Robot.Secret

		public := helpers.CreateImagePullSecretPublic(client.Namespace, withImagePullSecretSuffix(deployment.Name), registry, username, password)
		secret, err = client.CreateSecret(deployment.ID, appName, public)
		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	if service.NotCreated(ss.GetDeployment(appName)) {
		public := helpers.CreateMainAppDeploymentPublic(
			client.Namespace,
			deployment.Name,
			mainApp.Image,
			mainApp.InternalPort,
			params.Envs,
			params.Volumes,
			params.InitCommands,
			[]string{secret.Name},
		)
		_, err = client.CreateK8sDeployment(deployment.ID, appName, public)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	if service.NotCreated(ss.GetService(appName)) {
		public := helpers.CreateServicePublic(
			client.Namespace,
			deployment.Name,
			mainApp.InternalPort,
			mainApp.InternalPort,
		)
		_, err = client.CreateService(deployment.ID, appName, public)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress main
	if ingress := ss.GetIngress(appName); service.NotCreated(ingress) {
		var public *k8sModels.IngressPublic
		if params.Private {
			public = &k8sModels.IngressPublic{
				Placeholder: true,
			}
		} else {
			public = helpers.CreateIngressPublic(
				client.Namespace,
				deployment.Name,
				ss.GetService(appName).Name,
				ss.GetService(appName).Port,
				GetExternalFQDN(deployment.Name, client.Zone),
			)
		}

		_, err = client.CreateIngress(deployment.ID, appName, public)
		if err != nil {
			return makeError(err)
		}
	}

	// Ingress custom domain
	if mainApp.CustomDomain != nil && !params.Private {
		if ingress := ss.GetIngress(appNameCustomDomain); service.NotCreated(ingress) {
			public := helpers.CreateCustomDomainIngressPublic(
				client.Namespace,
				withCustomDomainSuffix(deployment.Name),
				ss.GetService(appName).Name,
				ss.GetService(appName).Port,
				*mainApp.CustomDomain,
			)

			_, err = client.CreateIngress(deployment.ID, appNameCustomDomain, public)
			if err != nil {
				if strings.Contains(err.Error(), "is already defined in ingress") {
					return makeError(CustomDomainInUseErr)
				}

				return makeError(err)
			}
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

	client, err := helpers.New(&deployment.Subsystems.K8s, deployment.Zone, getNamespaceName(deployment.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	// Ingress
	for mapName := range ss.IngressMap {
		err = client.DeleteIngress(deployment.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for mapName := range ss.ServiceMap {
		err = client.DeleteService(deployment.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for mapName := range ss.DeploymentMap {
		err = client.DeleteK8sDeployment(deployment.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolumeClaim
	for pvcName := range ss.PvcMap {
		err = client.DeletePVC(deployment.ID, pvcName)
		if err != nil {
			return makeError(err)
		}
	}

	// PersistentVolume
	for mapName := range ss.PvMap {
		err = client.DeletePV(deployment.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// Job
	for mapName := range ss.JobMap {
		err = client.DeleteJob(deployment.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for mapName := range ss.SecretMap {
		err = client.DeleteSecret(deployment.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	// Namespace
	err = client.DeleteNamespace(deployment.ID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Update(id string, params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s for deployment %s. details: %w", id, err)
	}

	if *params == (deploymentModel.UpdateParams{}) {
		return nil
	}

	deployment, err := deploymentModel.New().GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", id, "not found for k8s update assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&deployment.Subsystems.K8s, deployment.Zone, getNamespaceName(deployment.OwnerID))
	if err != nil {
		return makeError(err)
	}

	if params.InternalPort != nil {
		err = updateInternalPort(client, deployment, *params.InternalPort)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Envs != nil {
		err = updateEnvs(client, deployment, *params.Envs, params.InternalPort)
		if err != nil {
			return makeError(err)
		}
	}

	if params.CustomDomain != nil && (params.Private == nil || !*params.Private) {
		err = updateCustomDomain(client, deployment, *params.CustomDomain)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Private != nil {
		err = updatePrivate(client, deployment, *params.Private)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Volumes != nil {
		err = updateVolumes(client, deployment, *params.Volumes)
		if err != nil {
			return makeError(err)
		}
	}

	if params.Image != nil {
		err := updateImage(client, deployment, *params.Image)
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

	client, err := helpers.New(&deployment.Subsystems.K8s, deployment.Zone, getNamespaceName(deployment.OwnerID))
	if err != nil {
		return makeError(err)
	}

	err = client.SsClient.RestartDeployment(&k8sDeployment)
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

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		log.Println("main app not created when repairing k8s assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&deployment.Subsystems.K8s, deployment.Zone, getNamespaceName(deployment.OwnerID))
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	err = client.RepairNamespace(deployment.ID, func() *k8sModels.NamespacePublic {
		return helpers.CreateNamespacePublic(deployment.OwnerID)
	})

	if err != nil {
		return makeError(err)
	}

	for mapName := range ss.DeploymentMap {
		err = client.RepairK8sDeployment(deployment.ID, mapName, func() *k8sModels.DeploymentPublic {
			if mapName == appName {
				var pullSecrets []string
				if !deployment.Subsystems.Harbor.Project.Public {
					if service.Created(deployment.Subsystems.K8s.GetSecret(appName)) {
						log.Println("secret not created when repairing k8s assuming it was deleted")
						return nil
					} else {
						pullSecrets = []string{deployment.Subsystems.K8s.GetSecret(appName).Name}
					}
				}

				return helpers.CreateMainAppDeploymentPublic(
					ss.Namespace.FullName,
					deployment.Name,
					mainApp.Image,
					mainApp.InternalPort,
					mainApp.Envs,
					mainApp.Volumes,
					mainApp.InitCommands,
					pullSecrets,
				)
			}
			return nil
		})
		if err != nil {
			return makeError(err)
		}
	}

	for mapName := range ss.ServiceMap {
		err = client.RepairService(deployment.ID, mapName, func() *k8sModels.ServicePublic {
			if mapName == appName {
				return helpers.CreateServicePublic(
					client.Namespace,
					deployment.Name,
					mainApp.InternalPort,
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

	if mainApp.Private {
		for mapName := range ss.IngressMap {
			err = client.DeleteIngress(deployment.ID, mapName)
			if err != nil {
				return makeError(err)
			}
		}
	} else {
		k8sService := ss.GetService(appName)
		if !service.Created(k8sService) {
			log.Println("k8s service not created when recreating ingress. assuming it was deleted")
			return nil
		}

		if !service.Created(mainIngress) {
			ingressPublic := helpers.CreateIngressPublic(
				client.Namespace,
				deployment.Name,
				k8sService.Name,
				k8sService.Port,
				GetExternalFQDN(deployment.Name, client.Zone),
			)
			_, err = client.CreateIngress(deployment.ID, appName, ingressPublic)
			if err != nil {
				return makeError(err)
			}
		} else {
			err = client.RepairIngress(deployment.ID, appName, func() *k8sModels.IngressPublic {
				return helpers.CreateIngressPublic(
					client.Namespace,
					deployment.Name,
					k8sService.Name,
					k8sService.Port,
					GetExternalFQDN(deployment.Name, client.Zone),
				)
			})
			if err != nil {
				return makeError(err)
			}
		}

		if mainApp.CustomDomain != nil {
			if !service.Created(ss.GetIngress(appNameCustomDomain)) {
				ingressPublic := helpers.CreateCustomDomainIngressPublic(
					client.Namespace,
					withCustomDomainSuffix(deployment.Name),
					k8sService.Name,
					k8sService.Port,
					*mainApp.CustomDomain,
				)

				_, err = client.CreateIngress(deployment.ID, appNameCustomDomain, ingressPublic)
				if err != nil {
					if strings.Contains(err.Error(), "is already defined in ingress") {
						return makeError(CustomDomainInUseErr)
					}

					return makeError(err)
				}
			} else {
				err = client.RepairIngress(deployment.ID, appNameCustomDomain, func() *k8sModels.IngressPublic {
					return helpers.CreateCustomDomainIngressPublic(
						deployment.Subsystems.K8s.Namespace.FullName,
						withCustomDomainSuffix(deployment.Name),
						k8sService.Name,
						k8sService.Port,
						*mainApp.CustomDomain,
					)
				})
				if err != nil {
					if strings.Contains(err.Error(), "is already defined in ingress") {
						return makeError(CustomDomainInUseErr)
					}

					return makeError(err)
				}
			}
		}
	}

	imagePullSecret := ss.GetSecret(appNameImagePullSecret)
	if !service.Created(imagePullSecret) {
		registry := conf.Env.DockerRegistry.URL
		username := deployment.Subsystems.Harbor.Robot.HarborName
		password := deployment.Subsystems.Harbor.Robot.Secret

		public := helpers.CreateImagePullSecretPublic(client.Namespace, withImagePullSecretSuffix(deployment.Name), registry, username, password)
		_, err = client.CreateSecret(deployment.ID, appName, public)
		if err != nil {
			return makeError(err)
		}
	} else {
		err = client.RepairSecret(deployment.ID, appNameImagePullSecret, func() *k8sModels.SecretPublic {
			return helpers.CreateImagePullSecretPublic(
				deployment.Subsystems.K8s.Namespace.FullName,
				withImagePullSecretSuffix(deployment.Name),
				conf.Env.DockerRegistry.URL,
				deployment.Subsystems.Harbor.Robot.HarborName,
				deployment.Subsystems.Harbor.Robot.Secret,
			)
		})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func updateInternalPort(client *helpers.Client, deployment *deploymentModel.Deployment, internalPort int) error {
	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return nil
	}

	k8sService := client.K8s.GetService(mainApp.Name)
	if service.Created(k8sService) {
		if k8sService.Port != internalPort {
			k8sService.TargetPort = internalPort

			err := client.SsClient.UpdateService(k8sService)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func updateEnvs(client *helpers.Client, deployment *deploymentModel.Deployment, envs []deploymentModel.Env, internalPort *int) error {
	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return nil
	}

	k8sDeployment := client.K8s.GetDeployment(mainApp.Name)
	if service.Created(k8sDeployment) {
		var port int
		if internalPort != nil {
			port = *internalPort
		} else {
			port = mainApp.InternalPort
		}

		k8sEnvs := []k8sModels.EnvVar{
			{Name: "PORT", Value: strconv.Itoa(port)},
		}

		for _, env := range envs {
			if env.Name == "PORT" {
				continue
			}

			k8sEnvs = append(k8sEnvs, k8sModels.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			})
		}

		k8sDeployment.EnvVars = k8sEnvs

		err := client.SsClient.UpdateDeployment(k8sDeployment)
		if err != nil {
			return err
		}

		client.K8s.SetDeployment(mainApp.Name, *k8sDeployment)

		err = deploymentModel.New().UpdateSubsystemByID(deployment.ID, "k8s", "deploymentMap", &client.K8s.DeploymentMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateCustomDomain(client *helpers.Client, deployment *deploymentModel.Deployment, customDomain string) error {
	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return nil
	}

	k8sService := client.K8s.GetService(mainApp.Name)
	if service.Created(k8sService) {
		ingress := client.K8s.GetIngress(appNameCustomDomain)
		if service.Created(ingress) {
			ingress.CustomCert = &k8sModels.CustomCert{
				ClusterIssuer: "letsencrypt-prod-deploy-http",
				CommonName:    customDomain,
			}
			ingress.Hosts = []string{customDomain}
		} else {
			ingress = helpers.CreateCustomDomainIngressPublic(
				client.Namespace,
				withCustomDomainSuffix(deployment.Name),
				k8sService.Name,
				k8sService.Port,
				customDomain,
			)
		}

		err := client.RecreateIngress(deployment.ID, appNameCustomDomain, ingress)
		if err != nil {
			if strings.Contains(err.Error(), "is already defined in ingress") {
				return CustomDomainInUseErr
			}

			return err
		}
	} else {
		log.Println("k8s service", mainApp.Name, "not found for k8s update assuming it was deleted")
	}

	return nil
}

func updatePrivate(client *helpers.Client, deployment *deploymentModel.Deployment, private bool) error {
	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return nil
	}

	if private {
		for mapName := range client.K8s.IngressMap {
			err := client.DeleteIngress(deployment.ID, mapName)
			if err != nil {
				return err
			}
		}
	} else {
		if !service.Created(client.K8s.GetIngress(appName)) {
			_, err := client.CreateIngress(deployment.ID, appName, helpers.CreateIngressPublic(
				client.Namespace,
				deployment.Name,
				client.K8s.GetService(appName).Name,
				client.K8s.GetService(appName).Port,
				GetExternalFQDN(deployment.Name, client.Zone),
			))
			if err != nil {
				return err
			}
		}

		if mainApp.CustomDomain != nil {
			if !service.Created(client.K8s.GetIngress(appNameCustomDomain)) {
				_, err := client.CreateIngress(
					deployment.ID,
					withCustomDomainSuffix(deployment.Name),
					helpers.CreateCustomDomainIngressPublic(
						client.Namespace,
						withCustomDomainSuffix(deployment.Name),
						client.K8s.GetService(appName).Name,
						client.K8s.GetService(appName).Port,
						*mainApp.CustomDomain,
					),
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func updateVolumes(client *helpers.Client, deployment *deploymentModel.Deployment, volumes []deploymentModel.Volume) error {
	// delete deployment, pvcs and pvs
	// then
	// create new deployment, pvcs and pvs

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return nil
	}

	var pullSecrets []string
	if !deployment.Subsystems.Harbor.Project.Public {
		if service.NotCreated(deployment.Subsystems.K8s.GetSecret(appName)) {
			log.Println("k8s secret not created when updating volumes. assuming it was deleted")
			return nil
		} else {
			pullSecrets = []string{deployment.Subsystems.K8s.GetSecret(appName).Name}
		}
	}

	err := client.DeleteK8sDeployment(deployment.ID, appName)
	if err != nil {
		return err
	}

	for mapName, pvc := range client.K8s.PvcMap {
		err = client.DeletePVC(pvc.ID, mapName)
		if err != nil {
			return err
		}
	}

	for mapName, pv := range client.K8s.PvMap {
		err = client.DeletePV(pv.ID, mapName)
		if err != nil {
			return err
		}
	}

	// clear the maps
	client.K8s.DeploymentMap = make(map[string]k8sModels.DeploymentPublic)
	client.K8s.PvcMap = make(map[string]k8sModels.PvcPublic)
	client.K8s.PvMap = make(map[string]k8sModels.PvPublic)

	// since we depend on the namespace, we must ensure it is actually created here
	if !service.NotCreated(&client.K8s.Namespace) {
		public := helpers.CreateNamespacePublic(deployment.OwnerID)
		namespace, err := client.CreateNamespace(deployment.ID, public)
		if err != nil {
			return err
		}

		client.K8s.SetNamespace(*namespace)
	}

	for _, volume := range volumes {
		k8sName := fmt.Sprintf("%s-%s", deployment.Name, volume.Name)
		capacity := conf.Env.Deployment.Resources.Limits.Storage
		nfsPath := path.Join(client.Zone.Storage.NfsParentPath, deployment.OwnerID, "user", volume.ServerPath)

		pvPublic := helpers.CreatePvPublic(k8sName, capacity, nfsPath, client.Zone.Storage.NfsServer)
		_, err = client.CreatePV(deployment.ID, volume.Name, pvPublic)
		if err != nil {
			return err
		}

		pvcPublic := helpers.CreatePvcPublic(client.K8s.Namespace.FullName, k8sName, capacity, k8sName)
		_, err = client.CreatePVC(deployment.ID, volume.Name, pvcPublic)
		if err != nil {
			return err
		}
	}

	public := helpers.CreateMainAppDeploymentPublic(client.K8s.Namespace.FullName,
		deployment.Name,
		mainApp.Image,
		mainApp.InternalPort,
		mainApp.Envs,
		volumes,
		mainApp.InitCommands,
		pullSecrets,
	)
	_, err = client.CreateK8sDeployment(deployment.ID, appName, public)
	if err != nil {
		return err
	}

	return nil
}

func updateImage(client *helpers.Client, deployment *deploymentModel.Deployment, image string) error {
	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return nil
	}

	if image != mainApp.Image {
		oldPublic := client.K8s.GetDeployment(appName)
		if oldPublic.Created() {
			newPublic := oldPublic
			newPublic.Image = image

			err := client.SsClient.UpdateDeployment(newPublic)
			if err != nil {
				return err
			}

			client.K8s.SetDeployment(appName, *newPublic)

			err = deploymentModel.New().UpdateSubsystemByID(deployment.ID, "k8s", "deploymentMap", client.K8s.DeploymentMap)
		}
	}

	return nil
}
