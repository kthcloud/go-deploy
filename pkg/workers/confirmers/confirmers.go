package confirmers

import (
	"fmt"
	"go-deploy/models/sys/deployment"
	"go-deploy/models/sys/vm"
	"go-deploy/pkg/app"
	"log"
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
		return fmt.Errorf("failed to check if harbor is created for deployment %s. details: %s", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID != 0 &&
		harbor.Robot.ID != 0 &&
		harbor.Repository.ID != 0 &&
		harbor.Webhook.ID != 0, nil
}

func HarborDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor is created for deployment %s. details: %s", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID == 0 &&
		harbor.Robot.ID == 0 &&
		harbor.Repository.ID == 0 &&
		harbor.Webhook.ID == 0, nil
}

func CSCreated(vm *vm.VM) (bool, error) {
	cs := &vm.Subsystems.CS

	_, hasSshRule := cs.PortForwardingRuleMap["__ssh"]

	return cs.VM.ID != "" && hasSshRule, nil
}

func CSDeleted(vm *vm.VM) (bool, error) {
	cs := &vm.Subsystems.CS

	return cs.VM.ID == "" && len(cs.PortForwardingRuleMap) == 0, nil
}

func Setup(ctx *app.Context) {
	log.Println("starting confirmers")
	go deploymentConfirmer(ctx)
	go vmConfirmer(ctx)
}
