package confirm

import (
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/gpu"
	"go-deploy/models/sys/vm"
	"go-deploy/service"
)

func appCreatedK8s(deployment *deploymentModels.Deployment, app *deploymentModels.App) bool {
	for _, volume := range app.Volumes {
		pv, ok := deployment.Subsystems.K8s.GetPvMap()[deployment.Name+"-"+volume.Name]
		if !pv.Created() || !ok {
			return false
		}

		pvc, ok := deployment.Subsystems.K8s.GetPvcMap()[deployment.Name+"-"+volume.Name]
		if !pvc.Created() || !ok {
			return false
		}
	}

	deploymentCreated := false
	for mapName, k8sDeployment := range deployment.Subsystems.K8s.DeploymentMap {
		if k8sDeployment.Created() && mapName == deployment.Name {
			deploymentCreated = true
		}
	}

	serviceCreated := false
	for mapName, k8sService := range deployment.Subsystems.K8s.ServiceMap {
		if k8sService.Created() && mapName == deployment.Name {
			serviceCreated = true
		}
	}

	ingressCreated := false
	for mapName, ingress := range deployment.Subsystems.K8s.IngressMap {
		if ingress.Created() && mapName == deployment.Name {
			ingressCreated = true
		}
	}

	secretCreated := false
	if deployment.Type == deploymentModels.TypeCustom {
		for mapName, secret := range deployment.Subsystems.K8s.SecretMap {
			if secret.Created() && mapName == deployment.Name+"-image-pull-secret" {
				secretCreated = true
			}
		}
	} else {
		secretCreated = true
	}

	hpaCreated := false
	for mapName, hpa := range deployment.Subsystems.K8s.HpaMap {
		if hpa.Created() && mapName == deployment.Name {
			hpaCreated = true
		}
	}

	return deploymentCreated && serviceCreated && ingressCreated && secretCreated && hpaCreated
}

func appDeletedK8s(deployment *deploymentModels.Deployment, app *deploymentModels.App) bool {
	for _, volume := range app.Volumes {
		pv := deployment.Subsystems.K8s.PvMap[volume.Name]
		if pv.Created() {
			return false
		}

		pvc := deployment.Subsystems.K8s.PvcMap[volume.Name]
		if pvc.Created() {
			return false
		}
	}

	deploymentDeleted := true
	for mapName, k8sDeployment := range deployment.Subsystems.K8s.DeploymentMap {
		if k8sDeployment.Created() && mapName == deployment.Name {
			deploymentDeleted = false
		}
	}

	serviceDeleted := true
	for mapName, k8sService := range deployment.Subsystems.K8s.ServiceMap {
		if k8sService.Created() && mapName == deployment.Name {
			serviceDeleted = false
		}
	}

	ingressDeleted := true
	for mapName, ingress := range deployment.Subsystems.K8s.IngressMap {
		if ingress.Created() && mapName == deployment.Name {
			ingressDeleted = false
		}
	}

	secretDeleted := true
	if deployment.Type == deploymentModels.TypeCustom {
		for mapName, secret := range deployment.Subsystems.K8s.SecretMap {
			if secret.Created() && mapName == deployment.Name+"-image-pull-secret" {
				secretDeleted = false
			}
		}
	}

	hpaDeleted := true
	for mapName, hpa := range deployment.Subsystems.K8s.HpaMap {
		if hpa.Created() && mapName == deployment.Name {
			hpaDeleted = false
		}
	}

	return deploymentDeleted && serviceDeleted && ingressDeleted && secretDeleted && hpaDeleted
}

func k8sCreated(deployment *deploymentModels.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is created for deployment %s. details: %w", deployment.Name, err)
	}

	for _, app := range deployment.Apps {
		if !appCreatedK8s(deployment, &app) {
			return false, nil
		}
	}

	for mapName, secret := range deployment.Subsystems.K8s.SecretMap {
		if mapName == "wildcard-cert" && !secret.Created() {
			return false, nil
		}
	}

	k8s := &deployment.Subsystems.K8s
	return k8s.Namespace.Created(), nil
}

func k8sDeleted(deployment *deploymentModels.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is deleted for deployment %s. details: %w", deployment.Name, err)
	}

	for _, app := range deployment.Apps {
		if !appDeletedK8s(deployment, &app) {
			return false, nil
		}
	}

	for mapName, secret := range deployment.Subsystems.K8s.SecretMap {
		if mapName == "wildcard-cert" && secret.Created() {
			return false, nil
		}
	}

	k8s := &deployment.Subsystems.K8s
	return k8s.Namespace.ID == "", nil
}

func harborCreated(deployment *deploymentModels.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor is created for deployment %s. details: %w", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	if harbor.Placeholder {
		return true, nil
	}

	return harbor.Project.Created() &&
		harbor.Robot.Created() &&
		harbor.Repository.Created() &&
		harbor.Webhook.Created(), nil
}

func harborDeleted(deployment *deploymentModels.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor is created for deployment %s. details: %w", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID == 0 &&
		harbor.Robot.ID == 0 &&
		harbor.Repository.ID == 0 &&
		harbor.Webhook.ID == 0, nil
}

func gitHubCreated(deployment *deploymentModels.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if github is created for deployment %s. details: %w", deployment.Name, err)
	}

	github := &deployment.Subsystems.GitHub
	if github.Placeholder {
		return true, nil
	}

	return github.Webhook.Created(), nil
}

func gitHubDeleted(deployment *deploymentModels.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if github is created for deployment %s. details: %w", deployment.Name, err)
	}

	github := &deployment.Subsystems.GitHub

	if github.Placeholder {
		return true, nil
	}

	return github.Webhook.ID == 0, nil
}

func csCreated(vm *vm.VM) (bool, error) {
	cs := &vm.Subsystems.CS

	sshRule, ok := cs.PortForwardingRuleMap["__ssh"]
	if !ok {
		return false, nil
	}

	return cs.VM.Created() && sshRule.Created(), nil
}

func csDeleted(vm *vm.VM) (bool, error) {
	cs := &vm.Subsystems.CS

	for _, rule := range cs.PortForwardingRuleMap {
		if rule.Created() {
			return false, nil
		}
	}

	return cs.VM.ID == "", nil
}

func k8sCreatedVM(vm *vm.VM) (bool, error) {
	k8s := &vm.Subsystems.K8s

	portsWithHttpProxy := 0
	for _, port := range vm.Ports {
		if port.Name == "__ssh" {
			continue
		}

		if port.HttpProxy == nil {
			continue
		}

		portsWithHttpProxy++

		resourceName := vm.Name + "-" + port.Name

		if service.NotCreated(k8s.GetDeployment(resourceName)) {
			return false, nil
		}

		if service.NotCreated(k8s.GetService(resourceName)) {
			return false, nil
		}

		if service.NotCreated(k8s.GetIngress(resourceName)) {
			return false, nil
		}
	}

	if portsWithHttpProxy > 1 && !k8s.Namespace.Created() {
		return false, nil
	}

	return true, nil
}

func k8sDeletedVM(vm *vm.VM) (bool, error) {
	k8s := &vm.Subsystems.K8s

	if len(k8s.DeploymentMap) > 0 {
		return false, nil
	}

	if len(k8s.ServiceMap) > 0 {
		return false, nil
	}

	if len(k8s.IngressMap) > 0 {
		return false, nil
	}

	return k8s.Namespace.ID == "", nil
}

func gpuCleared(vm *vm.VM) (bool, error) {
	exists, err := gpu.New().WithVM(vm.ID).ExistsAny()
	if err != nil {
		return false, err
	}

	return !exists, nil
}
