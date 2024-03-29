package config

import (
	"go-deploy/models/model"
	"go-deploy/pkg/imp/kubevirt/kubevirt"
	"k8s.io/client-go/kubernetes"
)

// The following structs are used to parse the config.yaml file
// into a struct that can be used by the application.

type RancherConfigSource struct {
	ClusterID string `yaml:"clusterId"`
}

type CloudStackConfigSource struct {
	ClusterID   string `yaml:"clusterId"`
	ExternalURL string `yaml:"externalUrl"`
}

type DeploymentZone struct {
	Name       string `yaml:"name"`
	Namespaces struct {
		Deployment string `yaml:"deployment"`
		VM         string `yaml:"vm"`
		System     string `yaml:"system"`
	}
	Description             string `yaml:"description"`
	ParentDomain            string `yaml:"parentDomain"`
	ParentDomainVmHttpProxy string `yaml:"parentDomainVmHttpProxy"`
	CustomDomainIP          string `yaml:"customDomainIp"`
	NetworkPolicies         []struct {
		Name   string `yaml:"name"`
		Egress []struct {
			IP struct {
				Allow  string   `yaml:"allow"`
				Except []string `yaml:"except"`
			} `yaml:"ip"`
		} `yaml:"egress"`
	} `yaml:"networkPolicies"`
	IngressNamespace string `yaml:"ingressNamespace"`

	ConfigSource interface{} `yaml:"configSource"`
	Storage      struct {
		ParentDomain        string `yaml:"parentDomain"`
		NfsServer           string `yaml:"nfsServer"`
		NfsParentPath       string `yaml:"nfsParentPath"`
		VmStorageParentPath string `yaml:"vmStorageParentPath"`
	} `yaml:"storage"`

	K8sClient      *kubernetes.Clientset
	KubeVirtClient *kubevirt.Clientset

	// KubeVirt VM v2
	ParentDomainVM string `yaml:"parentDomainVm"`
	LoadBalancerIP string `yaml:"loadBalancerIp"`
	PortRange      struct {
		Start int `yaml:"start"`
		End   int `yaml:"end"`
	} `yaml:"portRange"`
}

type VmZone struct {
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
	RepairInterval                 int    `yaml:"repairInterval"`
	PingInterval                   int    `yaml:"pingInterval"`
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
	Zones []DeploymentZone `yaml:"zones"`
}

type VM struct {
	AdminSshPublicKey string   `yaml:"adminSshPublicKey"`
	RepairInterval    int      `yaml:"repairInterval"`
	Zones             []VmZone `yaml:"zones"`
}

type ConfigType struct {
	Version string `yaml:"version"`

	Port          int    `yaml:"port"`
	ExternalUrl   string `yaml:"externalUrl"`
	Manager       string `yaml:"manager"`
	SessionSecret string `yaml:"sessionSecret"`
	TestMode      bool   `yaml:"testMode"`

	GPU struct {
		PrivilegedGPUs []string `yaml:"privilegedGpus"`
		ExcludedHosts  []string `yaml:"excludedHosts"`
		ExcludedGPUs   []string `yaml:"excludedGpus"`
	} `yaml:"gpu_repo"`

	Registry struct {
		URL              string `yaml:"url"`
		PlaceholderImage string `yaml:"placeholderImage"`
		VmHttpProxyImage string `yaml:"vmHttpProxyImage"`
	} `yaml:"registry"`

	Deployment Deployment `yaml:"deployment"`
	VM         VM         `yaml:"vm"`

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

	CS struct {
		URL    string `yaml:"url"`
		ApiKey string `yaml:"apiKey"`
		Secret string `yaml:"secret"`
	} `yaml:"cs"`

	Rancher struct {
		URL    string `yaml:"url"`
		ApiKey string `yaml:"apiKey"`
		Secret string `yaml:"secret"`
	}

	Landing struct {
		URL      string `yaml:"url"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		ClientID string `yaml:"clientId"`
	} `yaml:"landing"`

	Harbor struct {
		URL           string `yaml:"url"`
		User          string `yaml:"user"`
		Password      string `yaml:"password"`
		WebhookSecret string `yaml:"webhookSecret"`
	} `yaml:"harbor"`

	GitLab struct {
		URL   string `yaml:"url"`
		Token string `yaml:"token"`
	}

	GitHub struct {
		DevClient struct {
			ID     string `yaml:"id"`
			Secret string `yaml:"secret"`
		} `yaml:"devClient"`
		ProdClient struct {
			ID     string `yaml:"id"`
			Secret string `yaml:"secret"`
		} `yaml:"prodClient"`
	}
}
