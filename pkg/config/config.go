package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/config"
	"go-deploy/models/version"
	"go-deploy/pkg/imp/kubevirt/kubevirt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/rancher"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"
	"time"
)

// Config is the global configuration object.
// It is populated from the config file, and managed as a singleton.
var Config config.ConfigType

// SetupEnvironment loads the configuration from the config file and sets up the environment.
func SetupEnvironment(mode string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to set up environment. details: %w", err)
	}

	filepath, found := os.LookupEnv("DEPLOY_CONFIG_FILE")
	if !found {
		return makeError(fmt.Errorf("config file not found. please set DEPLOY_CONFIG_FILE environment variable"))
	}

	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		return makeError(err)
	}

	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		return makeError(err)
	}

	Config.Mode = mode
	Config.Filepath = filepath
	config.LastRoleReload = time.Now()

	log.Println("go-deploy", version.AppVersion)

	// Fetch the roles from the config
	if len(Config.Roles) == 0 {
		log.Println("WARNING: no roles found in config")
	} else {
		var roles []string
		for _, role := range Config.Roles {
			roles = append(roles, role.Name)
		}
		log.Printf("Roles (in order): " + strings.Join(roles, "->"))
	}

	err = checkConfig()
	if err != nil {
		return makeError(err)
	}

	err = setupK8sClusters()
	if err != nil {
		return makeError(err)
	}

	err = validateConfig()
	if err != nil {
		return err
	}

	return nil
}

// checkConfig asserts that the config is correct.
func checkConfig() error {
	return nil
}

// setupK8sClusters sets up the k8s clusters.
func setupK8sClusters() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to set up k8s clusters. details: %w", err)
	}

	for idx, zone := range Config.Zones {
		sourceType, ok := zone.K8s.ConfigSource.(map[string]interface{})
		if !ok {
			return makeError(fmt.Errorf("failed to parse type of config source for zone %s", zone.Name))
		}

		configType, ok := sourceType["type"].(string)
		if !ok {
			return makeError(fmt.Errorf("failed to parse type of config source for zone %s", zone.Name))
		}

		log.Printf("Setting up K8s cluster for zone %s (%d/%d)", zone.Name, idx+1, len(Config.Zones))

		switch configType {
		case "localPath":
			{
				var zoneConfig config.LocalPathConfigSource
				err := mapstructure.Decode(sourceType, &zoneConfig)
				if err != nil {
					return makeError(fmt.Errorf("failed to parse file config source for zone %s. details: %w", zone.Name, err))
				}

				k8sClient, kubevirtClient, err := createClientFromLocalPathConfig(zone.Name, &zoneConfig)
				if err != nil {
					return makeError(err)
				}

				Config.Zones[idx].K8s.Client = k8sClient
				Config.Zones[idx].K8s.KubeVirtClient = kubevirtClient
			}
		case "rancher":
			{
				var zoneConfig config.RancherConfigSource
				err := mapstructure.Decode(sourceType, &zoneConfig)
				if err != nil {
					return makeError(fmt.Errorf("failed to parse rancher config source for zone %s. details: %w", zone.Name, err))
				}

				k8sClient, kubevirtClient, err := createClientFromRancherConfig(zone.Name, &zoneConfig)
				if err != nil {
					return makeError(err)
				}

				Config.Zones[idx].K8s.Client = k8sClient
				Config.Zones[idx].K8s.KubeVirtClient = kubevirtClient
			}
		}
	}

	log.Println("K8s clusters setup done")
	return nil
}

// createClientFromLocalPathConfig creates a k8s client from a local path config.
func createClientFromLocalPathConfig(zoneName string, config *config.LocalPathConfigSource) (*kubernetes.Clientset, *kubevirt.Clientset, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s client from local path config (zone: %s). details: %w", zoneName, err)
	}

	kubeConfig, err := os.ReadFile(config.Path)
	if err != nil {
		return nil, nil, makeError(err)
	}

	return createK8sClients(kubeConfig)
}

// createClientFromRancherConfig creates a k8s client from a rancher config.
func createClientFromRancherConfig(zoneName string, config *config.RancherConfigSource) (*kubernetes.Clientset, *kubevirt.Clientset, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s client from rancher config (zone: %s). details: %w", zoneName, err)
	}

	// Check cache, if cache/rancher-rancher-cluster-name exists, use it
	// If not, create a new client and save it to cache
	if _, err := os.Stat(fmt.Sprintf("cache/rancher-%s.config", config.ClusterName)); os.IsNotExist(err) {
		rancherClient, err := rancher.New(&rancher.ClientConf{
			URL:    config.URL,
			ApiKey: config.ApiKey,
			Secret: config.Secret,
		})
		if err != nil {
			return nil, nil, makeError(err)
		}

		kubeConfig, err := rancherClient.ReadClusterKubeConfig(config.ClusterName)
		if err != nil {
			return nil, nil, makeError(err)
		}

		if kubeConfig == "" {
			return nil, nil, makeError(fmt.Errorf("kubeconfig not found for cluster %s", config.ClusterName))
		}

		cacheDir := "cache"
		if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
			err = os.Mkdir(cacheDir, 0755)
			if err != nil {
				return nil, nil, makeError(err)
			}
		}

		err = os.WriteFile(fmt.Sprintf("cache/rancher-%s.config", config.ClusterName), []byte(kubeConfig), 0644)
		if err != nil {
			return nil, nil, makeError(err)
		}

		return createK8sClients([]byte(kubeConfig))
	}

	kubeConfig, err := os.ReadFile(fmt.Sprintf("cache/rancher-%s.config", config.ClusterName))
	if err != nil {
		return nil, nil, makeError(err)
	}

	return createK8sClients(kubeConfig)
}

// createK8sClients creates a k8s client from config data.
func createK8sClients(configData []byte) (*kubernetes.Clientset, *kubevirt.Clientset, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	kubeConfig, err := clientcmd.RESTConfigFromKubeConfig(configData)
	if err != nil {
		return nil, nil, makeError(err)
	}

	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, makeError(err)
	}

	kubeVirtClient, err := kubevirt.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, makeError(err)
	}

	return k8sClient, kubeVirtClient, nil
}

// validateConfig validates the config and throws an error if it is invalid.
// It is only concerned with static validation, and does not check for dynamic issues.
func validateConfig() error {
	// Ensure there is a default zone
	foundDefaultDeployment := false
	foundDefaultVM := false
	for _, zone := range Config.Zones {
		if zone.Name == Config.VM.DefaultZone {
			foundDefaultVM = true
		}

		if zone.Name == Config.Deployment.DefaultZone {
			foundDefaultDeployment = true
		}
	}

	if !foundDefaultDeployment {
		return fmt.Errorf("no default deployment zone found")
	}

	if !foundDefaultVM {
		return fmt.Errorf("no default VM zone found")
	}

	return nil
}
