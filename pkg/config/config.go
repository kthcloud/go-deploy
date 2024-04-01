package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/config"
	"go-deploy/pkg/imp/cloudstack"
	"go-deploy/pkg/imp/kubevirt/kubevirt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/rancher"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"regexp"
	"strings"
)

// Config is the global configuration object.
// It is populated from the config file, and managed as a singleton.
var Config config.ConfigType

// SetupEnvironment loads the configuration from the config file and sets up the environment.
func SetupEnvironment() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup environment. details: %w", err)
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

	log.Println("go-deploy", Config.Version)

	if len(Config.Roles) == 0 {
		log.Println("WARNING: no roles found in config")
	} else {
		var roles []string
		for idx, role := range Config.Roles {
			if idx == len(Config.Roles)-1 {
				log.Printf("%s", role.Name)
				break
			}
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

	return nil
}

// checkConfig asserts that the config is correct.
// This includes checking for unique zone names.
func checkConfig() error {
	uniqueNames := make(map[string]bool)
	for _, zone := range Config.Deployment.Zones {
		if uniqueNames[zone.Name] {
			return fmt.Errorf("found duplicate deployment zone name: %s", zone.Name)
		}
		uniqueNames[zone.Name] = true
	}

	uniqueNames = make(map[string]bool)
	for _, zone := range Config.VM.Zones {
		if uniqueNames[zone.Name] {
			return fmt.Errorf("found duplicate vm zone name: %s", zone.Name)
		}
		uniqueNames[zone.Name] = true
	}

	return nil
}

// setupK8sClusters sets up the k8s clusters.
func setupK8sClusters() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s clusters. details: %w", err)
	}

	for idx, zone := range Config.Deployment.Zones {
		sourceType, ok := zone.ConfigSource.(map[string]interface{})
		if !ok {
			return makeError(fmt.Errorf("failed to parse type of config source for zone %s", zone.Name))
		}

		configType, ok := sourceType["type"].(string)
		if !ok {
			return makeError(fmt.Errorf("failed to parse type of config source for zone %s", zone.Name))
		}

		log.Printf(" - Setting up k8s cluster for zone %s (%d/%d)", zone.Name, idx+1, len(Config.Deployment.Zones))

		switch configType {
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

				Config.Deployment.Zones[idx].K8sClient = k8sClient
				Config.Deployment.Zones[idx].KubeVirtClient = kubevirtClient
			}
		case "cloudstack":
			{
				var zoneConfig config.CloudStackConfigSource
				err := mapstructure.Decode(sourceType, &zoneConfig)
				if err != nil {
					return makeError(fmt.Errorf("failed to parse cloudstack config source for zone %s. details: %w", zone.Name, err))
				}

				k8sClient, kubevirtClient, err := createClientFromCloudStackConfig(zone.Name, &zoneConfig)
				if err != nil {
					return makeError(err)
				}

				Config.Deployment.Zones[idx].K8sClient = k8sClient
				Config.Deployment.Zones[idx].KubeVirtClient = kubevirtClient
			}
		}
	}

	log.Println("k8s clusters setup done")
	return nil
}

// createClientFromRancherConfig creates a k8s client from a rancher config.
func createClientFromRancherConfig(name string, config *config.RancherConfigSource) (*kubernetes.Clientset, *kubevirt.Clientset, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s client from rancher config. details: %w", err)
	}

	rancherClient, err := rancher.New(&rancher.ClientConf{
		URL:    Config.Rancher.URL,
		ApiKey: Config.Rancher.ApiKey,
		Secret: Config.Rancher.Secret,
	})
	if err != nil {
		return nil, nil, makeError(err)
	}

	kubeConfig, err := rancherClient.ReadClusterKubeConfig(config.ClusterID)
	if err != nil {
		return nil, nil, makeError(err)
	}

	if kubeConfig == "" {
		return nil, nil, makeError(fmt.Errorf("kubeconfig not found for cluster %s", config.ClusterID))
	}

	return createK8sClients([]byte(kubeConfig))
}

// createClientFromCloudStackConfig creates a k8s client from a cloudstack config.
func createClientFromCloudStackConfig(name string, config *config.CloudStackConfigSource) (*kubernetes.Clientset, *kubevirt.Clientset, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s client from cloudstack config. details: %w", err)
	}

	csClient := cloudstack.NewAsyncClient(
		Config.CS.URL,
		Config.CS.ApiKey,
		Config.CS.Secret,
		true,
	)

	listClusterParams := csClient.Kubernetes.NewListKubernetesClustersParams()
	listClusterParams.SetListall(true)
	listClusterParams.SetId(config.ClusterID)
	clusters, err := csClient.Kubernetes.ListKubernetesClusters(listClusterParams)
	if err != nil {
		return nil, nil, makeError(err)
	}

	if len(clusters.KubernetesClusters) == 0 {
		return nil, nil, makeError(fmt.Errorf("cluster with name %s not found", name))
	}

	if len(clusters.KubernetesClusters) > 1 {
		return nil, nil, makeError(fmt.Errorf("multiple clusters found for name %s", name))
	}

	params := csClient.Kubernetes.NewGetKubernetesClusterConfigParams()
	params.SetId(clusters.KubernetesClusters[0].Id)

	clusterConfig, err := csClient.Kubernetes.GetKubernetesClusterConfig(params)
	if err != nil {
		return nil, nil, makeError(err)
	}

	// use regex to replace the private ip in config.ConfigData 172.31.1.* with the public ip
	regex := regexp.MustCompile(`https://172.[0-9]+.[0-9]+.[0-9]+:6443`)
	clusterConfig.Configdata = regex.ReplaceAllString(clusterConfig.Configdata, config.ExternalURL)

	return createK8sClients([]byte(clusterConfig.Configdata))
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
