package confirm

import (
	"errors"
	"fmt"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/vm_port_repo"
	"net"
)

// appDeletedK8s checks if the K8s setup for the given app is deleted.
func appDeletedK8s(deployment *model.Deployment, app *model.App) bool {
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
	if deployment.Type == model.DeploymentTypeCustom {
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
func k8sDeletedDeployment(deployment *model.Deployment) (bool, error) {
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
	return !k8s.Namespace.Created(), nil
}

// harborDeleted checks if the Harbor setup for the given deployment is deleted.
func harborDeleted(deployment *model.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor is created for deployment %s. details: %w", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID == 0 &&
		harbor.Robot.ID == 0 &&
		harbor.Repository.ID == 0 &&
		harbor.Webhook.ID == 0, nil
}

// k8sDeleted checks if the K8s setup for the given storage manager is deleted.
func k8sDeletedSM(sm *model.SM) (bool, error) {
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

	if len(k8s.SecretMap) > 0 {
		return false, nil
	}

	return !k8s.Namespace.Created(), nil
}

// k8sDeleted checks if the K8s setup for the given VM is deleted.
func k8sDeletedVM(vm *model.VM) (bool, error) {
	k8s := &vm.Subsystems.K8s

	if len(k8s.DeploymentMap) > 0 {
		return false, nil
	}

	if k8s.VM.Created() {
		return false, nil
	}

	if len(k8s.ServiceMap) > 0 {
		return false, nil
	}

	if len(k8s.IngressMap) > 0 {
		return false, nil
	}

	if len(k8s.PvcMap) > 0 {
		return false, nil
	}

	if len(k8s.PvMap) > 0 {
		return false, nil
	}

	if len(k8s.VmSnapshotMap) > 0 {
		return false, nil
	}

	return !k8s.Namespace.Created(), nil
}

// portsCleared checks if the ports setup for the given VM is cleared.
func portsCleared(vm *model.VM) (bool, error) {
	exists, err := vm_port_repo.New().WithVmID(vm.ID).ExistsAny()
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// checkCustomDomain checks if the custom domain is setup.
// This polls the TXT record of the custom domain to check if the secret is set.
// It returns (exists, match, txtRecord, error).
func checkCustomDomain(domain string, secret string) (bool, bool, string, error) {
	subDomain := config.Config.Deployment.CustomDomainTxtRecordSubdomain

	txtRecordDomain := subDomain + "." + domain
	txtRecord, err := net.LookupTXT(txtRecordDomain)
	if err != nil {
		// If error is "no such host", it means the DNS record does not exist yet
		var targetErr *net.DNSError
		if ok := errors.As(err, &targetErr); ok && targetErr.IsNotFound {
			return false, false, "", nil
		}

		return false, false, "", fmt.Errorf("failed to lookup TXT record under %s for custom domain %s. details: %w", subDomain, domain, err)
	}

	exists := len(txtRecord) > 0
	if !exists {
		return false, false, "", nil
	}

	match := false
	for _, r := range txtRecord {
		if r == secret {
			match = true
			break
		}
	}

	return exists, match, txtRecord[0], nil
}
