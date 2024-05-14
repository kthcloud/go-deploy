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
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"slices"
	"time"
)

type K8sGenerator struct {
	generators.K8sGeneratorBase

	vm     *model.VM
	zone   *configModels.Zone
	client *k8s.Client

	namespace           string
	extraAuthorizedKeys []string
}

type CloudInit struct {
	FQDN            string          `yaml:"fqdn"`
	Users           []CloudInitUser `yaml:"users"`
	SshPasswordAuth bool            `yaml:"ssh_pwauth"`
	RunCMD          []string        `yaml:"runcmd"`
}

type CloudInitUser struct {
	Name              string   `yaml:"name"`
	Sudo              []string `yaml:"sudo"`
	Passwd            string   `yaml:"passwd,omitempty"`
	LockPasswd        bool     `yaml:"lock_passwd"`
	Shell             string   `yaml:"shell"`
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

func K8s(vm *model.VM, zone *configModels.Zone, client *k8s.Client, namespace string, extraAuthorizedKeys []string) *K8sGenerator {
	return &K8sGenerator{
		vm:                  vm,
		zone:                zone,
		client:              client,
		namespace:           namespace,
		extraAuthorizedKeys: extraAuthorizedKeys,
	}
}

func (kg *K8sGenerator) Namespace() *models.NamespacePublic {
	ns := models.NamespacePublic{
		Name: kg.namespace,
	}

	if n := &kg.vm.Subsystems.K8s.Namespace; subsystems.Created(n) {
		ns.CreatedAt = n.CreatedAt
	}

	return &ns
}

func (kg *K8sGenerator) VMs() []models.VmPublic {
	sshPublicKeys := make([]string, len(kg.extraAuthorizedKeys)+1)
	sshPublicKeys[0] = kg.vm.SshPublicKey
	copy(sshPublicKeys[1:], kg.extraAuthorizedKeys)

	cloudInit := CloudInit{
		FQDN: kg.vm.Name,
		Users: []CloudInitUser{
			{
				Name:              "root",
				Sudo:              []string{"ALL=(ALL) NOPASSWD:ALL"},
				LockPasswd:        false,
				Shell:             "/bin/bash",
				SshAuthorizedKeys: sshPublicKeys,
			},
		},
		SshPasswordAuth: false,
		RunCMD:          []string{"git clone https://github.com/kthcloud/boostrap-vm.git init && cd init && chmod +x run.sh && ./run.sh"},
	}

	vmPublic := models.VmPublic{
		Name:      vmName(kg.vm),
		Namespace: kg.namespace,
		Labels:    map[string]string{"owner-id": kg.vm.OwnerID},
		CpuCores:  kg.vm.Specs.CpuCores,
		RAM:       kg.vm.Specs.RAM,
		DiskSize:  kg.vm.Specs.DiskSize,
		GPUs:      make([]string, 0),
		CloudInit: createCloudInitString(&cloudInit),
		// Temporary image URL
		Image:     config.Config.VM.Image,
		Running:   true,
		CreatedAt: time.Time{},
	}

	if vm := kg.vm.Subsystems.K8s.GetVM(vmName(kg.vm)); subsystems.Created(vm) {
		vmPublic.ID = vm.ID
		vmPublic.Running = vm.Running
		vmPublic.CreatedAt = vm.CreatedAt
	}

	return []models.VmPublic{vmPublic}
}

func (kg *K8sGenerator) Services() []models.ServicePublic {
	res := make([]models.ServicePublic, 0)

	portMap := kg.vm.PortMap

	for _, port := range portMap {
		res = append(res, models.ServicePublic{
			Name:      vmServiceName(kg.vm, vmPfrName(port.Port, port.Protocol)),
			Namespace: kg.namespace,
			Ports: []models.Port{{
				Name:       vmPfrName(port.Port, port.Protocol),
				Protocol:   port.Protocol,
				Port:       0, // This is set externally
				TargetPort: port.Port,
			}},
			Selector: map[string]string{
				keys.LabelDeployName: vmName(kg.vm),
			},
			LoadBalancerIP: kg.zone.K8s.LoadBalancerIP,
		})

		if port.HttpProxy != nil {
			res = append(res, models.ServicePublic{
				Name:      vmProxyServiceName(kg.vm, vmPfrName(port.Port, port.Protocol)),
				Namespace: kg.namespace,
				Ports: []models.Port{{
					Name:       vmPfrName(port.Port, port.Protocol),
					Protocol:   "tcp",
					Port:       8080,
					TargetPort: port.Port,
				}},
				Selector: map[string]string{
					keys.LabelDeployName: vmName(kg.vm),
				},
			})
		}
	}

	for mapName, s := range kg.vm.Subsystems.K8s.GetServiceMap() {
		idx := slices.IndexFunc(res, func(service models.ServicePublic) bool {
			return service.Name == mapName
		})
		if idx != -1 {
			// Copy over the external ports
			for servicePortIdx, servicePort := range res[idx].Ports {
				if servicePort.Port != 0 {
					continue
				}

				portIdx := slices.IndexFunc(s.Ports, func(port models.Port) bool {
					return port.Name == servicePort.Name
				})
				if portIdx != -1 {
					res[idx].Ports[servicePortIdx].Port = s.Ports[portIdx].Port
				}
			}

			res[idx].CreatedAt = s.CreatedAt
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
			ServiceName:  vmProxyServiceName(kg.vm, vmPfrName(port.Port, port.Protocol)),
			ServicePort:  8080,
			IngressClass: config.Config.Deployment.IngressClass,
			Hosts:        []string{vmProxyExternalURL(port.HttpProxy.Name, kg.zone)},
			TlsSecret:    &tlsSecret,
			CustomCert:   nil,
			Placeholder:  false,
		})
		if port.HttpProxy.CustomDomain != nil && port.HttpProxy.CustomDomain.Status == model.CustomDomainStatusActive {
			res = append(res, models.IngressPublic{
				Name:         vmProxyCustomDomainIngressName(kg.vm, port.HttpProxy.Name),
				Namespace:    kg.namespace,
				ServiceName:  vmProxyServiceName(kg.vm, vmPfrName(port.Port, port.Protocol)),
				ServicePort:  8080,
				IngressClass: config.Config.Deployment.IngressClass,
				Hosts:        []string{port.HttpProxy.CustomDomain.Domain},
				Placeholder:  false,
				CustomCert: &models.CustomCert{
					ClusterIssuer: kg.zone.K8s.ClusterIssuer,
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

func (kg *K8sGenerator) NetworkPolicies() []models.NetworkPolicyPublic {
	res := make([]models.NetworkPolicyPublic, 0)

	if !anyHttpProxy(kg.vm) {
		return nil
	}

	for _, egressRule := range kg.zone.NetworkPolicies {
		egressRules := make([]models.EgressRule, 0)
		for _, egress := range egressRule.Egress {
			egressRules = append(egressRules, models.EgressRule{
				IpBlock: &models.IpBlock{
					CIDR:   egress.IP.Allow,
					Except: egress.IP.Except,
				},
			})
		}

		np := models.NetworkPolicyPublic{
			Name:        vmNetworkPolicyName(kg.vm.Name, egressRule.Name),
			Namespace:   kg.namespace,
			Selector:    map[string]string{keys.LabelDeployName: kg.vm.Name},
			EgressRules: egressRules,
			IngressRules: []models.IngressRule{
				{
					PodSelector:       map[string]string{"owner-id": kg.vm.OwnerID},
					NamespaceSelector: map[string]string{"kubernetes.io/metadata.name": kg.namespace},
				},
				{
					NamespaceSelector: map[string]string{"kubernetes.io/metadata.name": kg.zone.K8s.IngressNamespace},
				},
			},
		}

		if npo := kg.vm.Subsystems.K8s.GetNetworkPolicy(egressRule.Name); subsystems.Created(npo) {
			np.CreatedAt = npo.CreatedAt
		}

		res = append(res, np)
	}

	return res
}

// createCloudInitString creates a cloud-init string from a cloud-init struct
func createCloudInitString(cloudInit *CloudInit) string {
	// Convert cloud-init struct to yaml
	yamlBytes, err := yaml.Marshal(cloudInit)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to marshal cloud-init struct to yaml. details: %w", err))
		return ""
	}

	return "#cloud-config\n" + string(yamlBytes)
}

// vmName returns the VM name for a VM
func vmName(vm *model.VM) string {
	return vm.Name
}

// vmServiceName returns the service name for a VM.
// portName should be created using the vmPfrName function.
func vmServiceName(vm *model.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vmProxyServiceName returns the service name for a VM proxy.
// portName should be created using the vmPfrName function.
func vmProxyServiceName(vm *model.VM, portName string) string {
	return fmt.Sprintf("%s-%s-proxy", vm.Name, portName)
}

// vmPfrName is a helper function to create a name for a port forwarding rule.
// It is to ensure that there are no restrictions on the name, while still being able to identify it
func vmPfrName(privatePort int, protocol string, suffix ...string) string {
	return fmt.Sprintf("priv-%d-prot-%s", privatePort, protocol)
}

// vmProxyIngressName returns the ingress name for a VM proxy
func vmProxyIngressName(vm *model.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vmProxyCustomDomainIngressName returns the ingress name for a VM proxy custom domain
func vmProxyCustomDomainIngressName(vm *model.VM, portName string) string {
	return fmt.Sprintf("%s-%s-custom-domain", vm.Name, portName)
}

// vmProxyExternalURL returns the external URL for a VM proxy
func vmProxyExternalURL(portName string, zone *configModels.Zone) string {
	return fmt.Sprintf("%s.%s", portName, zone.Domains.ParentVmApp)
}

// vmNetworkPolicyName returns the network policy name for a VM or Deployment
func vmNetworkPolicyName(name, egressRuleName string) string {
	return fmt.Sprintf("%s-%s", name, egressRuleName)
}

// anyHttpProxy returns true if a VM has any HTTP proxy ports
func anyHttpProxy(vm *model.VM) bool {
	for _, port := range vm.PortMap {
		if port.HttpProxy != nil {
			return true
		}
	}

	return false
}
