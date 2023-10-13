package confirm

import (
	"fmt"
	"go-deploy/models/sys/deployment"
	"go-deploy/models/sys/vm"
)

func appCreatedK8s(deployment *deployment.Deployment, app *deployment.App) bool {
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
	for mapName, service := range deployment.Subsystems.K8s.ServiceMap {
		if service.Created() && mapName == deployment.Name {
			serviceCreated = true
		}
	}

	ingressCreated := false
	for mapName, ingress := range deployment.Subsystems.K8s.IngressMap {
		if ingress.Created() && mapName == deployment.Name {
			ingressCreated = true
		}
	}

	return deploymentCreated && serviceCreated && ingressCreated
}

func appDeletedK8s(deployment *deployment.Deployment, app *deployment.App) bool {
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
	for mapName, service := range deployment.Subsystems.K8s.ServiceMap {
		if service.Created() && mapName == deployment.Name {
			serviceDeleted = false
		}
	}

	ingressDeleted := true
	for mapName, ingress := range deployment.Subsystems.K8s.IngressMap {
		if ingress.Created() && mapName == deployment.Name {
			ingressDeleted = false
		}
	}

	return deploymentDeleted && serviceDeleted && ingressDeleted
}

func k8sCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is created for deployment %s. details: %w", deployment.Name, err)
	}

	for _, app := range deployment.Apps {
		if !appCreatedK8s(deployment, &app) {
			return false, nil
		}
	}

	k8s := &deployment.Subsystems.K8s
	return k8s.Namespace.Created(), nil
}

func k8sDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is deleted for deployment %s. details: %w", deployment.Name, err)
	}

	for _, app := range deployment.Apps {
		if !appDeletedK8s(deployment, &app) {
			return false, nil
		}
	}

	k8s := &deployment.Subsystems.K8s
	return k8s.Namespace.ID == "", nil
}

func harborCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor is created for deployment %s. details: %w", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	if harbor.Placeholder {
		return true, nil
	}

	return harbor.Project.ID != 0 &&
		harbor.Robot.ID != 0 &&
		harbor.Repository.ID != 0 &&
		harbor.Webhook.ID != 0, nil
}

func harborDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor is created for deployment %s. details: %w", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID == 0 &&
		harbor.Robot.ID == 0 &&
		harbor.Repository.ID == 0 &&
		harbor.Webhook.ID == 0, nil
}

func gitHubCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if github is created for deployment %s. details: %w", deployment.Name, err)
	}

	github := &deployment.Subsystems.GitHub
	if github.Placeholder {
		return true, nil
	}

	return github.Webhook.ID != 0, nil
}

func gitHubDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if github is created for deployment %s. details: %w", deployment.Name, err)
	}

	github := &deployment.Subsystems.GitHub

	return !github.Placeholder &&
		github.Webhook.ID == 0, nil
}

func csCreated(vm *vm.VM) (bool, error) {
	cs := &vm.Subsystems.CS

	_, hasSshRule := cs.PortForwardingRuleMap["__ssh"]

	return cs.VM.ID != "" && hasSshRule, nil
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

func gpuCleared(vm *vm.VM) (bool, error) {
	return vm.GpuID == "", nil
}
