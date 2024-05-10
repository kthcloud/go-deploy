package resources

import (
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	"go-deploy/service/generators"
	"go-deploy/utils"
	v1 "k8s.io/api/core/v1"
)

type K8sGenerator struct {
	generators.K8sGeneratorBase

	vm             *model.VM
	vmZone         *configModels.LegacyZone
	deploymentZone *configModels.Zone
	client         *k8s.Client

	namespace string
}

func K8s(vm *model.VM, vmZone *configModels.LegacyZone, deploymentZone *configModels.Zone, client *k8s.Client, namespace string) *K8sGenerator {
	return &K8sGenerator{
		vm:             vm,
		vmZone:         vmZone,
		deploymentZone: deploymentZone,
		client:         client,
		namespace:      namespace,
	}
}

func (kg *K8sGenerator) Namespace() *models.NamespacePublic {
	createNamespace := false
	for _, port := range kg.vm.PortMap {
		if port.HttpProxy != nil {
			createNamespace = true
			break
		}
	}

	if !createNamespace {
		return nil
	}

	ns := models.NamespacePublic{
		Name: kg.namespace,
	}

	if n := &kg.vm.Subsystems.K8s.Namespace; subsystems.Created(n) {
		ns.CreatedAt = n.CreatedAt
	}

	return &ns
}

func (kg *K8sGenerator) Deployments() []models.DeploymentPublic {
	res := make([]models.DeploymentPublic, 0)

	portMap := kg.vm.PortMap

	for _, port := range portMap {
		if port.HttpProxy == nil {
			continue
		}

		csPort := kg.vm.Subsystems.CS.GetPortForwardingRule(vmPfrName(port.Port, port.Protocol))
		if csPort == nil {
			continue
		}

		envVars := []models.EnvVar{
			{Name: "PORT", Value: "8080"},
			{Name: "VM_PORT", Value: fmt.Sprintf("%d", csPort.PublicPort)},
			{Name: "URL", Value: vmProxyExternalURL(port.HttpProxy.Name, kg.deploymentZone)},
			{Name: "VM_URL", Value: kg.vmZone.ParentDomain},
		}

		res = append(res, models.DeploymentPublic{
			Name:             vmProxyDeploymentName(kg.vm, port.HttpProxy.Name),
			Namespace:        kg.namespace,
			Labels:           map[string]string{"owner-id": kg.vm.OwnerID},
			Image:            config.Config.Registry.VmHttpProxyImage,
			ImagePullSecrets: make([]string, 0),
			EnvVars:          envVars,
			Resources: models.Resources{
				Limits: models.Limits{
					CPU:    fmt.Sprintf("%f", config.Config.Deployment.Resources.Limits.CPU),
					Memory: fmt.Sprintf("%fGi", config.Config.Deployment.Resources.Limits.RAM),
				},
				Requests: models.Requests{
					CPU:    fmt.Sprintf("%f", config.Config.Deployment.Resources.Requests.CPU),
					Memory: fmt.Sprintf("%fGi", config.Config.Deployment.Resources.Requests.RAM),
				},
			},
			Command:        make([]string, 0),
			Args:           make([]string, 0),
			InitCommands:   make([]string, 0),
			InitContainers: make([]models.InitContainer, 0),
			Volumes:        make([]models.Volume, 0),
		})
	}

	for mapName, k8sDeployment := range kg.vm.Subsystems.K8s.GetDeploymentMap() {
		idx := 0
		matchedIdx := -1
		for _, port := range portMap {
			if port.HttpProxy == nil {
				continue
			}

			if vmProxyDeploymentName(kg.vm, port.HttpProxy.Name) == mapName {
				matchedIdx = idx
				break
			}

			idx++
		}

		if matchedIdx != -1 {
			res[idx].CreatedAt = k8sDeployment.CreatedAt
		}
	}

	return res
}

func (kg *K8sGenerator) Services() []models.ServicePublic {
	res := make([]models.ServicePublic, 0)

	portMap := kg.vm.PortMap

	for _, port := range portMap {
		if port.HttpProxy == nil {
			continue
		}

		res = append(res, models.ServicePublic{
			Name:      vmProxyServiceName(kg.vm, port.HttpProxy.Name),
			Namespace: kg.namespace,
			Ports:     []models.Port{{Name: vmPfrName(port.Port, port.Protocol), Protocol: port.Protocol, Port: 8080, TargetPort: 8080}},
			Selector: map[string]string{
				keys.LabelDeployName: vmProxyDeploymentName(kg.vm, port.HttpProxy.Name),
			},
		})
	}

	for mapName, svc := range kg.vm.Subsystems.K8s.GetServiceMap() {
		idx := 0
		matchedIdx := -1
		for _, port := range portMap {
			if port.HttpProxy == nil {
				continue
			}

			if vmProxyServiceName(kg.vm, port.HttpProxy.Name) == mapName {
				matchedIdx = idx
				break
			}

			idx++
		}

		if matchedIdx != -1 {
			res[idx].CreatedAt = svc.CreatedAt
		}
	}

	return res
}

func (kg *K8sGenerator) Ingresses() []models.IngressPublic {
	res := make([]models.IngressPublic, 0)

	portMap := kg.vm.PortMap

	for _, port := range portMap {
		if port.HttpProxy == nil {
			continue
		}

		tlsSecret := constants.WildcardCertSecretName
		res = append(res, models.IngressPublic{
			Name:         vmProxyIngressName(kg.vm, port.HttpProxy.Name),
			Namespace:    kg.namespace,
			ServiceName:  vmProxyServiceName(kg.vm, port.HttpProxy.Name),
			ServicePort:  8080,
			IngressClass: config.Config.Deployment.IngressClass,
			Hosts:        []string{vmProxyExternalURL(port.HttpProxy.Name, kg.deploymentZone)},
			TlsSecret:    &tlsSecret,
			CustomCert:   nil,
			Placeholder:  false,
		})
		if port.HttpProxy.CustomDomain != nil && port.HttpProxy.CustomDomain.Status == model.CustomDomainStatusActive {
			res = append(res, models.IngressPublic{
				Name:         vmProxyCustomDomainIngressName(kg.vm, port.HttpProxy.Name),
				Namespace:    kg.namespace,
				ServiceName:  vmProxyServiceName(kg.vm, port.HttpProxy.Name),
				ServicePort:  8080,
				IngressClass: config.Config.Deployment.IngressClass,
				Hosts:        []string{port.HttpProxy.CustomDomain.Domain},
				Placeholder:  false,
				CustomCert: &models.CustomCert{
					ClusterIssuer: kg.deploymentZone.K8s.ClusterIssuer,
					CommonName:    port.HttpProxy.CustomDomain.Domain,
				},
				TlsSecret: nil,
			})
		}
	}

	for mapName, ingress := range kg.vm.Subsystems.K8s.GetIngressMap() {
		idx := 0
		matchedIdx := -1
		for _, port := range portMap {
			if port.HttpProxy == nil {
				continue
			}

			if vmProxyIngressName(kg.vm, port.HttpProxy.Name) == mapName ||
				(vmProxyCustomDomainIngressName(kg.vm, port.HttpProxy.Name) == mapName && port.HttpProxy.CustomDomain != nil) {
				matchedIdx = idx
				break
			}
		}

		if matchedIdx != -1 {
			res[idx].CreatedAt = ingress.CreatedAt
		}
	}

	return res
}

func (kg *K8sGenerator) Secrets() []models.SecretPublic {
	res := make([]models.SecretPublic, 0)

	createWildcardSecret := false
	for _, port := range kg.vm.PortMap {
		if port.HttpProxy != nil {
			createWildcardSecret = true
			break
		}
	}

	if !createWildcardSecret {
		return nil
	}

	// wildcard certificate
	/// swap namespaces temporarily
	var wildcardCertSecret *models.SecretPublic

	kg.client.Namespace = config.Config.Deployment.WildcardCertSecretNamespace
	defer func() { kg.client.Namespace = kg.namespace }()

	copyFrom, err := kg.client.ReadSecret(config.Config.Deployment.WildcardCertSecretName)
	if err != nil || copyFrom == nil {
		utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", config.Config.Deployment.WildcardCertSecretNamespace, config.Config.Deployment.WildcardCertSecretName, err))
	} else {
		wildcardCertSecret = &models.SecretPublic{
			Name:      constants.WildcardCertSecretName,
			Namespace: kg.namespace,
			Type:      string(v1.SecretTypeOpaque),
			Data:      copyFrom.Data,
		}
	}

	if secret := kg.vm.Subsystems.K8s.GetSecret(constants.WildcardCertSecretName); subsystems.Created(secret) {
		if wildcardCertSecret == nil {
			wildcardCertSecret = secret
		} else {
			wildcardCertSecret.CreatedAt = secret.CreatedAt
		}
	}

	if wildcardCertSecret != nil {
		res = append(res, *wildcardCertSecret)
	}

	return res
}

// vmProxyDeploymentName returns the deployment name for a VM proxy
func vmProxyDeploymentName(vm *model.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vmProxyServiceName returns the service name for a VM proxy
func vmProxyServiceName(vm *model.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vmProxyIngressName returns the ingress name for a VM proxy
func vmProxyIngressName(vm *model.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vmPfrName is a helper function to create a name for a PortForwardingRule.
// It is to ensure that there are no restrictions on the name, while still being able to identify it
func vmPfrName(privatePort int, protocol string) string {
	return fmt.Sprintf("priv-%d-prot-%s", privatePort, protocol)
}

// vmProxyCustomDomainIngressName returns the ingress name for a VM proxy custom domain
func vmProxyCustomDomainIngressName(vm *model.VM, portName string) string {
	return fmt.Sprintf("%s-%s-custom-domain", vm.Name, portName)
}

// vmProxyExternalURL returns the external URL for a VM proxy
func vmProxyExternalURL(portName string, zone *configModels.Zone) string {
	return fmt.Sprintf("%s.%s", portName, zone.Domains.ParentVmApp)
}
