package config

import (
	"go-deploy/models/model"
	"go-deploy/pkg/imp/kubevirt/kubevirt"
	"k8s.io/client-go/kubernetes"
	"time"
)

// The following structs are used to parse the config.yaml file
// into a struct that can be used by the application.

const (
	ZoneCapabilityDeployment = "deployment"
	ZoneCapabilityVM         = "vm"
)

type ConfigType struct {
	Version string `yaml:"version"`

	Port        int    `yaml:"port"`
	ExternalUrl string `yaml:"externalUrl"`
	Mode        string `yaml:"mode"`

	Deployment Deployment `yaml:"deployment"`
	VM         VM         `yaml:"vm"`

	Zones []Zone `yaml:"zones"`

	Timer struct {
		GpuSynchronize            time.Duration `yaml:"gpuSynchronize"`
		GpuLeaseSynchronize       time.Duration `yaml:"gpuLeaseSynchronize"`
		VmStatusUpdate            time.Duration `yaml:"vmStatusUpdate"`
		VmSnapshotUpdate          time.Duration `yaml:"vmSnapshotUpdate"`
		DeploymentStatusUpdate    time.Duration `yaml:"deploymentStatusUpdate"`
		DeploymentPingUpdate      time.Duration `yaml:"deploymentPingUpdate"`
		Snapshot                  time.Duration `yaml:"snapshot"`
		DeploymentRepair          time.Duration `yaml:"deploymentRepair"`
		VmRepair                  time.Duration `yaml:"vmRepair"`
		SmRepair                  time.Duration `yaml:"smRepair"`
		MetricsUpdate             time.Duration `yaml:"metricsUpdate"`
		JobFetch                  time.Duration `yaml:"jobFetch"`
		FailedJobFetch            time.Duration `yaml:"failedJobFetch"`
		DeploymentDeletionConfirm time.Duration `yaml:"deploymentDeletionConfirm"`
		VmDeletionConfirm         time.Duration `yaml:"vmDeletionConfirm"`
		SmDeletionConfirm         time.Duration `yaml:"smDeletionConfirm"`
		CustomDomainConfirm       time.Duration `yaml:"customDomainConfirm"`
	}

	GPU struct {
		PrivilegedGPUs []string `yaml:"privilegedGpus"`
		ExcludedHosts  []string `yaml:"excludedHosts"`
		ExcludedGPUs   []string `yaml:"excludedGpus"`
	} `yaml:"gpu"`

	Registry struct {
		URL              string `yaml:"url"`
		PlaceholderImage string `yaml:"placeholderImage"`
		VmHttpProxyImage string `yaml:"vmHttpProxyImage"`
	} `yaml:"registry"`

	Roles []model.Role `yaml:"roles"`

	Metrics struct {
		Interval int `yaml:"interval"`
	} `yaml:"metrics"`

	Keycloak struct {
		Url           string `yaml:"url"`
		Realm         string `yaml:"realm"`
		AdminGroup    string `yaml:"adminGroup"`
		StorageClient struct {
			ClientID     string `yaml:"clientId"`
			ClientSecret string `yaml:"clientSecret"`
		} `yaml:"storageClient"`
	} `yaml:"keycloak"`

	MongoDB struct {
		URL  string `yaml:"url"`
		Name string `yaml:"name"`
	} `yaml:"mongodb"`

	Redis struct {
		URL      string `yaml:"url"`
		Password string `yaml:"password"`
	}

	// CS is the CloudStack configuration
	// Deprecated: CloudStack will no longer be used
	CS struct {
		URL    string `yaml:"url"`
		ApiKey string `yaml:"apiKey"`
		Secret string `yaml:"secret"`
	} `yaml:"cs"`

	SysApi struct {
		URL      string `yaml:"url"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		ClientID string `yaml:"clientId"`
		// UseMock is a flag that indicates whether the sys-api should be mocked
		// This is useful for testing purposes, since the sys-api cannot be run locally
		UseMock bool `yaml:"useMock"`
	} `yaml:"sys-api"`

	Harbor struct {
		URL           string `yaml:"url"`
		User          string `yaml:"user"`
		Password      string `yaml:"password"`
		WebhookSecret string `yaml:"webhookSecret"`
	} `yaml:"harbor"`
}

type RancherConfigSource struct {
	ClusterName string `yaml:"clusterName"`
	URL         string `yaml:"url"`
	ApiKey      string `yaml:"apiKey"`
	Secret      string `yaml:"secret"`
}

type KubeConfigSource struct {
	Filepath string `yaml:"filepath"`
}

// CloudStackConfigSource
// Deprecated: Use RancherConfigSource or KubeConfigSource instead
type CloudStackConfigSource struct {
	ClusterID   string `yaml:"clusterId"`
	ExternalURL string `yaml:"externalUrl"`
}

type VM struct {
	AdminSshPublicKey string       `yaml:"adminSshPublicKey"`
	Zones             []LegacyZone `yaml:"zones"`
}

// LegacyZone is a zone that is used for VM v1
// Deprecated: User Zone instead
type LegacyZone struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	ParentDomain string `yaml:"parentDomain"`
	PortRange    struct {
		Start int `yaml:"start"`
		End   int `yaml:"end"`
	} `yaml:"portRange"`

	// CloudStack IDs
	ZoneID      string `yaml:"zoneId"`
	ProjectID   string `yaml:"projectId"`
	NetworkID   string `yaml:"networkId"`
	IpAddressID string `yaml:"ipAddressId"`
	TemplateID  string `yaml:"templateId"`
}

type Deployment struct {
	Port                           int    `yaml:"port"`
	Prefix                         string `yaml:"prefix"`
	WildcardCertSecretNamespace    string `yaml:"wildcardCertSecretNamespace"`
	WildcardCertSecretName         string `yaml:"wildcardCertSecretName"`
	CustomDomainTxtRecordSubdomain string `yaml:"customDomainTxtRecordSubdomain"`
	IngressClass                   string `yaml:"ingressClass"`
	Resources                      struct {
		AutoScale struct {
			CpuThreshold    int `yaml:"cpuThreshold"`
			MemoryThreshold int `yaml:"memoryThreshold"`
		} `yaml:"autoScale"`
		Limits struct {
			CPU     string `yaml:"cpu"`
			Memory  string `yaml:"memory"`
			Storage string `yaml:"storage"`
		} `yaml:"limits"`
		Requests struct {
			CPU    string `yaml:"cpu"`
			Memory string `yaml:"memory"`
		} `yaml:"requests"`
	} `yaml:"resources"`
}

type Zone struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	K8s struct {
		Namespaces struct {
			Deployment string `yaml:"deployment"`
			VM         string `yaml:"vm"`
			System     string `yaml:"system"`
		} `yaml:"namespaces"`
		// ConfigSource is the source of the Kubernetes configuration
		// It can be either RancherConfigSource or KubeConfigSource
		ConfigSource interface{} `yaml:"configSource"`
		// IngressNamespace is the namespace where the ingress resources are created, e.g. "ingress-nginx"
		IngressNamespace string `yaml:"ingressNamespace"`
		// LoadBalancerIP is the IP of the load balancer that is used for the ingress resources
		LoadBalancerIP string `yaml:"loadBalancerIp"`
		// Client is the Kubernetes client for the zone created by querying the ConfigSource
		Client *kubernetes.Clientset
		// KubeVirtClient is the KubeVirt client for the zone created by querying the ConfigSource
		KubeVirtClient *kubevirt.Clientset
	}

	Capabilities []string `yaml:"capabilities"`
	Domains      struct {
		ParentDeployment string `yaml:"parentDeployment"`
		ParentVM         string `yaml:"parentVm"`
		ParentVmApp      string `yaml:"parentVmApp"`
		ParentSM         string `yaml:"parentSm"`
	}
	Storage struct {
		NfsServer string `yaml:"nfsServer"`
		Paths     struct {
			ParentDeployment string `yaml:"parentDeployment"`
			ParentVM         string `yaml:"parentVm"`
		} `yaml:"paths"`
	} `yaml:"storage"`
	NetworkPolicies []struct {
		Name   string `yaml:"name"`
		Egress []struct {
			IP struct {
				Allow  string   `yaml:"allow"`
				Except []string `yaml:"except"`
			} `yaml:"ip"`
		} `yaml:"egress"`
	} `yaml:"networkPolicies"`
	PortRange struct {
		Start int `yaml:"start"`
		End   int `yaml:"end"`
	} `yaml:"portRange"`
}
