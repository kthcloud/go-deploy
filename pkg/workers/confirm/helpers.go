package confirm

import (
	"fmt"
	"go-deploy/models/sys/deployment"
	"go-deploy/models/sys/vm"
)

func k8sCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is created for deployment %s. details: %s", deployment.Name, err)
	}

	k8s := &deployment.Subsystems.K8s
	return k8s.Namespace.Name != "" &&
		k8s.Deployment.ID != "" &&
		k8s.Service.ID != "" &&
		k8s.Ingress.ID != "" || k8s.Ingress.Placeholder, nil
}

func k8sDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is deleted for deployment %s. details: %s", deployment.Name, err)
	}

	k8s := &deployment.Subsystems.K8s
	return k8s.Namespace.Name == "" &&
		k8s.Deployment.ID == "" &&
		k8s.Service.ID == "" &&
		k8s.Ingress.ID == "" && !k8s.Ingress.Placeholder, nil
}

func harborCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor is created for deployment %s. details: %s", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID != 0 &&
		harbor.Robot.ID != 0 &&
		harbor.Repository.ID != 0 &&
		harbor.Webhook.ID != 0, nil
}

func harborDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor is created for deployment %s. details: %s", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID == 0 &&
		harbor.Robot.ID == 0 &&
		harbor.Repository.ID == 0 &&
		harbor.Webhook.ID == 0, nil
}

func gitHubCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if github is created for deployment %s. details: %s", deployment.Name, err)
	}

	github := &deployment.Subsystems.GitHub
	if github.Placeholder {
		return true, nil
	}

	return github.Webhook.ID != 0, nil
}

func gitHubDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if github is created for deployment %s. details: %s", deployment.Name, err)
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

	return cs.VM.ID == "" && len(cs.PortForwardingRuleMap) == 0, nil
}

func gpuCleared(vm *vm.VM) (bool, error) {
	return vm.GpuID == "", nil
}
