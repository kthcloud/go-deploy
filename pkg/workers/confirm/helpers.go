package confirm

import (
	"errors"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/gpu"
	"go-deploy/models/sys/sm"
	"go-deploy/models/sys/vm"
	"go-deploy/models/sys/vm_port"
	"go-deploy/pkg/config"
	"go-deploy/utils"
	"log"
	"net"
)

// appDeletedK8s checks if the K8s setup for the given app is deleted.
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

// k8sDeleted checks if the K8s setup for the given deployment is deleted.
func k8sDeletedDeployment(deployment *deploymentModels.Deployment) (bool, error) {
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

// harborDeleted checks if the Harbor setup for the given deployment is deleted.
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

// gitHubDeleted checks if the GitHub setup for the given deployment is deleted.
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

// k8sDeleted checks if the K8s setup for the given storage manager is deleted.
func k8sDeletedSM(sm *sm.SM) (bool, error) {
	k8s := &sm.Subsystems.K8s

	if len(k8s.DeploymentMap) > 0 {
		return false, nil
	}

	if len(k8s.ServiceMap) > 0 {
		return false, nil
	}

	if len(k8s.IngressMap) > 0 {
		return false, nil
	}

	if len(k8s.PvMap) > 0 {
		return false, nil
	}

	if len(k8s.PvcMap) > 0 {
		return false, nil
	}

	if len(k8s.JobMap) > 0 {
		return false, nil
	}

	if len(k8s.SecretMap) > 0 {
		return false, nil
	}

	return k8s.Namespace.ID == "", nil
}

// csDeleted checks if the CS setup for the given VM is deleted.
func csDeleted(vm *vm.VM) (bool, error) {
	cs := &vm.Subsystems.CS

	for _, rule := range cs.PortForwardingRuleMap {
		if rule.Created() {
			return false, nil
		}
	}

	if cs.VM.Created() {
		return false, nil
	}

	return true, nil
}

// k8sDeleted checks if the K8s setup for the given VM is deleted.
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

// gpuCleared checks if the GPU setup for the given VM is cleared.
func gpuCleared(vm *vm.VM) (bool, error) {
	exists, err := gpu.New().WithVM(vm.ID).ExistsAny()
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// portsCleared checks if the ports setup for the given VM is cleared.
func portsCleared(vm *vm.VM) (bool, error) {
	exists, err := vm_port.New().WithVmID(vm.ID).ExistsAny()
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// checkCustomDomain checks if the custom domain is setup.
// This polls the TXT record of the custom domain to check if the secret is set.
func checkCustomDomain(domain string, secret string) (bool, error) {
	subDomain := config.Config.Deployment.CustomDomainTxtRecordSubdomain

	txtRecordDomain := subDomain + "." + domain
	txtRecord, err := net.LookupTXT(txtRecordDomain)
	if err != nil {
		// If error is "no such host", it means the DNS record does not exist yet
		var targetErr *net.DNSError
		if ok := errors.As(err, &targetErr); ok && targetErr.IsNotFound {
			log.Printf("no TXT record found under %s when confirming custom domain %s\n", subDomain, domain)
			return false, nil
		}

		utils.PrettyPrintError(fmt.Errorf("failed to lookup TXT record under %s for custom domain %s. details: %w", subDomain, domain, err))
		return false, err
	}

	if len(txtRecord) == 0 {
		log.Printf("no TXT record found under %s when confirming custom domain %s\n", subDomain, domain)
		return false, nil
	}

	match := false
	for _, r := range txtRecord {
		if r == secret {
			match = true
			break
		}
	}

	if !match {
		log.Printf("TXT record found under %s for custom domain %s but secret does not match\n", subDomain, domain)
		return false, nil
	}

	return match, nil

}
