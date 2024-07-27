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
	Port        int    `yaml:"port"`
	ExternalUrl string `yaml:"externalUrl"`
	// Mode is the mode in which the application is running
	// It is set using the command line flag --mode
	Mode string
	// Filepath is the path to the configuration file
	// It is set by the SetupEnvironment function when the configuration is loaded
	Filepath string

	Deployment Deployment `yaml:"deployment"`
	VM         VM         `yaml:"vm"`
	Discovery  struct {
		Token string `yaml:"token"`
	} `yaml:"discovery"`

	Zones []Zone `yaml:"zones"`

	Timer struct {
		DeploymentStatusUpdate    time.Duration `yaml:"deploymentStatusUpdate"`
		DeploymentPingUpdate      time.Duration `yaml:"deploymentPingUpdate"`
		DeploymentRepair          time.Duration `yaml:"deploymentRepair"`
		DeploymentDeletionConfirm time.Duration `yaml:"deploymentDeletionConfirm"`

		SmRepair          time.Duration `yaml:"smRepair"`
		SmDeletionConfirm time.Duration `yaml:"smDeletionConfirm"`

		VmStatusUpdate    time.Duration `yaml:"vmStatusUpdate"`
		VmRepair          time.Duration `yaml:"vmRepair"`
		VmDeletionConfirm time.Duration `yaml:"vmDeletionConfirm"`

		GpuSynchronize      time.Duration `yaml:"gpuSynchronize"`
		GpuLeaseSynchronize time.Duration `yaml:"gpuLeaseSynchronize"`

		CustomDomainConfirm  time.Duration `yaml:"customDomainConfirm"`
		StaleResourceCleanup time.Duration `yaml:"staleResourceCleanup"`

		MetricsUpdate time.Duration `yaml:"metricsUpdate"`

		JobFetch       time.Duration `yaml:"jobFetch"`
		FailedJobFetch time.Duration `yaml:"failedJobFetch"`

		FetchSystemStats      time.Duration `yaml:"fetchSystemStats"`
		FetchSystemCapacities time.Duration `yaml:"fetchSystemCapacities"`
		FetchSystemStatus     time.Duration `yaml:"fetchSystemStatus"`
		FetchSystemGpuInfo    time.Duration `yaml:"fetchSystemGpuInfo"`
	}

	GPU struct {
		PrivilegedGPUs []string `yaml:"privilegedGpus"`
		ExcludedHosts  []string `yaml:"excludedHosts"`
		ExcludedGPUs   []string `yaml:"excludedGpus"`
		AddMock        bool     `yaml:"addMock"`
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
		Url        string `yaml:"url"`
		Realm      string `yaml:"realm"`
		AdminGroup string `yaml:"adminGroup"`
		UserClient struct {
			ClientID     string `yaml:"clientId"`
			ClientSecret string `yaml:"clientSecret"`
		} `yaml:"userClient"`
	} `yaml:"keycloak"`

	MongoDB struct {
		URL  string `yaml:"url"`
		Name string `yaml:"name"`
	} `yaml:"mongodb"`

	Redis struct {
		URL      string `yaml:"url"`
		Password string `yaml:"password"`
	}

	Harbor struct {
		URL           string `yaml:"url"`
		User          string `yaml:"user"`
		Password      string `yaml:"password"`
		WebhookSecret string `yaml:"webhookSecret"`
	} `yaml:"harbor"`
}

type LocalPathConfigSource struct {
	Path string `yaml:"path"`
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

type VM struct {
	DefaultZone       string        `yaml:"defaultZone"`
	Lifetime          time.Duration `yaml:"lifetime"`
	AdminSshPublicKey string        `yaml:"adminSshPublicKey"`
	Image             string        `yaml:"image"`
}

type Deployment struct {
	DefaultZone string        `yaml:"defaultZone"`
	Port        int           `yaml:"port"`
	Prefix      string        `yaml:"prefix"`
	Lifetime    time.Duration `yaml:"lifetime"`
	Fallback    struct {
		// Disabled is the name of the fallback deployment that other deployments can reference
		// The fallback deployment can be created with go-deploy, or it can be created manually
		Disabled struct {
			// Name is the name of the fallback deployment
			Name string `yaml:"name"`
		} `yaml:"disabled"`
	} `yaml:"fallback"`

	WildcardCertSecretNamespace    string `yaml:"wildcardCertSecretNamespace"`
	WildcardCertSecretName         string `yaml:"wildcardCertSecretName"`
	CustomDomainTxtRecordSubdomain string `yaml:"customDomainTxtRecordSubdomain"`

	IngressClass string `yaml:"ingressClass"`

	Resources struct {
		AutoScale struct {
			CpuThreshold    int `yaml:"cpuThreshold"`
			MemoryThreshold int `yaml:"memoryThreshold"`
		} `yaml:"autoScale"`
		Limits struct {
			// CPU in cores (0.5 for 500m)
			CPU float64 `yaml:"cpu"`
			// RAM in GB (0.5 for 500Mi)
			RAM float64 `yaml:"memory"`
			// Storage in GB (1 for 1Gi, no decimal)
			Storage int `yaml:"storage"`
		} `yaml:"limits"`
		Requests struct {
			// CPU in cores (0.5 for 500m)
			CPU float64 `yaml:"cpu"`
			// RAM in GB (0.5 for 500Mi)
			RAM float64 `yaml:"memory"`
		} `yaml:"requests"`
	} `yaml:"resources"`
}

type Zone struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Enabled     bool   `yaml:"enabled"`

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
		// It is only set if the cluster sets dynamic load balancer IPs to ensure that the IP is always the same
		LoadBalancerIP *string `yaml:"loadBalancerIp"`
		ClusterIssuer  string  `yaml:"clusterIssuer"`
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
