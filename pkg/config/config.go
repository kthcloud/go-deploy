package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/config"
	"go-deploy/pkg/imp/cloudstack"
	"go-deploy/pkg/imp/kubevirt/kubevirt"
	"go-deploy/pkg/subsystems/rancher"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"regexp"
)

// Config is the global configuration object.
// It is populated from the config file, and managed as a singleton.
var Config config.ConfigType

// SetupEnvironment loads the configuration from the config file and sets up the environment.
func SetupEnvironment() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup environment. details: %w", err)
	}

	filepath, found := os.LookupEnv("DEPLOY_CONFIG_FILE")
	if !found {
		log.Fatalln(makeError(fmt.Errorf("config file not found. please set DEPLOY_CONFIG_FILE environment variable")))
	}

	log.Println("reading config from", filepath)
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf(makeError(err).Error())
	}

	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		log.Fatalf(makeError(err).Error())
	}

	log.Println("go-deploy", Config.Version)

	log.Println("loaded", len(Config.Roles), "roles in order:")
	for _, role := range Config.Roles {
		log.Println("\t", role.Name)
	}

	assertCorrectConfig()

	err = setupK8sClusters()
	if err != nil {
		log.Fatalln(makeError(err))
	}

	log.Println("config loading finished")
}

// assertCorrectConfig asserts that the config is correct.
func assertCorrectConfig() {
	uniqueNames := make(map[string]bool)
	for _, zone := range Config.Deployment.Zones {
		if uniqueNames[zone.Name] {
			log.Fatalln("deployment zone names must be unique")
		}
		uniqueNames[zone.Name] = true
	}

	uniqueNames = make(map[string]bool)
	for _, zone := range Config.VM.Zones {
		if uniqueNames[zone.Name] {
			log.Fatalln("vm zone names must be unique")
		}
		uniqueNames[zone.Name] = true
	}

	log.Println("config checks passed")
}

// setupK8sClusters sets up the k8s clusters.
func setupK8sClusters() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s clusters. details: %w", err)
	}

	for idx, zone := range Config.Deployment.Zones {
		sourceType, ok := zone.ConfigSource.(map[string]interface{})
		if !ok {
			log.Fatalln("failed to parse type of config source for zone", zone.Name)
		}

		configType, ok := sourceType["type"].(string)
		if !ok {
			log.Fatalln("failed to parse type of config source for zone", zone.Name)
		}

		switch configType {
		case "rancher":
			{
				var zoneConfig config.RancherConfigSource
				err := mapstructure.Decode(sourceType, &zoneConfig)
				if err != nil {
					log.Fatalln("failed to parse rancher config source for zone", zone.Name)
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
					log.Fatalln("failed to parse cloudstack config source for zone", zone.Name)
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

	log.Println("fetching k8s cluster for deployment zone", name)

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

	log.Println("fetching k8s cluster for deployment zone", name)

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
		log.Fatalln(makeError(err))
	}

	if len(clusters.KubernetesClusters) == 0 {
		log.Fatalln("cluster for deployment zone", name, "not found")
	}

	if len(clusters.KubernetesClusters) > 1 {
		log.Fatalln("multiple clusters for deployment zone", name, "found")
	}

	params := csClient.Kubernetes.NewGetKubernetesClusterConfigParams()
	params.SetId(clusters.KubernetesClusters[0].Id)

	clusterConfig, err := csClient.Kubernetes.GetKubernetesClusterConfig(params)
	if err != nil {
		log.Fatalln(makeError(err))
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
