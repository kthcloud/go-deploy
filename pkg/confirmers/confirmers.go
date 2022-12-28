package confirmers

import (
	"fmt"
	"go-deploy/models/deployment"
	"go-deploy/models/vm"
)

func NPMCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if npm setup is created for deployment %s. details: %s", deployment.Name, err)
	}

	return deployment.Subsystems.Npm.ProxyHost.ID != 0, nil
}

func NPMDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if npm setup is deleted for deployment %s. details: %s", deployment.Name, err)
	}

	return deployment.Subsystems.Npm.ProxyHost.ID == 0, nil
}

func K8sCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is created for deployment %s. details: %s", deployment.Name, err)
	}

	k8s := &deployment.Subsystems.K8s
	return k8s.Namespace.Name != "" &&
		k8s.Deployment.ID != "" &&
		k8s.Service.ID != "", nil
}

func K8sDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is deleted for deployment %s. details: %s", deployment.Name, err)
	}

	k8s := &deployment.Subsystems.K8s
	return k8s.Namespace.Name == "" &&
		k8s.Deployment.ID == "" &&
		k8s.Service.ID == "", nil
}

func HarborCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor setup is created for deployment %s. details: %s", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID != 0 &&
		harbor.Robot.ID != 0 &&
		harbor.Repository.ID != 0 &&
		harbor.Webhook.ID != 0, nil
}

func HarborDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor setup is created for deployment %s. details: %s", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID == 0 &&
		harbor.Robot.ID == 0 &&
		harbor.Repository.ID == 0 &&
		harbor.Webhook.ID == 0, nil
}

func CSCreated(vm *vm.VM) (bool, error) {
	cs := &vm.Subsystems.CS
	return len(cs.VM.ID) != 0, nil
}

func CSDeleted(vm *vm.VM) (bool, error) {
	cs := &vm.Subsystems.CS
	return len(cs.VM.ID) == 0, nil
}
