package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go-deploy/models/config"
	"go-deploy/pkg/imp/cloudstack"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"regexp"
)

var Config config.ConfigType

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
		case "cloudstack":
			{
				var zoneConfig config.CloudStackConfigSource
				err := mapstructure.Decode(sourceType, &zoneConfig)
				if err != nil {
					log.Fatalln("failed to parse cloudstack config source for zone", zone.Name)
				}

				client, err := createClientFromCloudStackConfig(zone.Name, &zoneConfig)
				if err != nil {
					return makeError(err)
				}

				Config.Deployment.Zones[idx].Client = client
			}
		}
	}

	log.Println("k8s clusters setup done")
	return nil
}

func createClientFromCloudStackConfig(name string, config *config.CloudStackConfigSource) (*kubernetes.Clientset, error) {
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
	regex := regexp.MustCompile(`https://172.31.1.[0-9]+:6443`)
	clusterConfig.Configdata = regex.ReplaceAllString(clusterConfig.Configdata, config.ExternalURL)

	return createK8sClient([]byte(clusterConfig.Configdata))
}

func createK8sClient(configData []byte) (*kubernetes.Clientset, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	kubeConfig, err := clientcmd.RESTConfigFromKubeConfig(configData)
	if err != nil {
		return nil, makeError(err)
	}

	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, makeError(err)
	}

	return k8sClient, nil
}
