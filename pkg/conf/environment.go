package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Environment struct {
	Port          int    `yaml:"port"`
	ExternalUrl   string `yaml:"externalUrl"`
	Manager       string `yaml:"manager"`
	SessionSecret string `yaml:"sessionSecret"`

	GpuConfFilepath string `yaml:"gpuMetaFilepath"`

	DockerRegistry struct {
		Url         string `yaml:"url"`
		PlaceHolder struct {
			Project    string `yaml:"project"`
			Repository string `yaml:"repository"`
		} `yaml:"placeHolder"`
	} `yaml:"dockerRegistry"`

	App struct {
		ParentDomain string `yaml:"parentDomain"`
		Port         int    `yaml:"port"`
		Prefix       string `yaml:"prefix"`
		DefaultQuota int    `yaml:"defaultQuota"`
	} `yaml:"app"`

	VM struct {
		ParentDomain      string `yaml:"parentDomain"`
		DefaultQuota      int    `yaml:"defaultQuota"`
		AdminSshPublicKey string `yaml:"adminSshPublicKey"`
	} `yaml:"vm"`

	Keycloak struct {
		Url        string `yaml:"url"`
		Realm      string `yaml:"realm"`
		AdminGroup string `yaml:"adminGroup"`
		GpuGroup   string `yaml:"gpuGroup"`
	} `yaml:"keycloak"`

	DB struct {
		Url  string `yaml:"url"`
		Name string `yaml:"name"`
	} `yaml:"db"`

	CS struct {
		Url    string `yaml:"url"`
		ApiKey string `yaml:"apiKey"`
		Secret string `yaml:"secret"`
	} `yaml:"cs"`

	PfSense struct {
		User      string `yaml:"user"`
		Password  string `yaml:"password"`
		Url       string `yaml:"url"`
		PublicIP  string `yaml:"publicIp"`
		PortRange struct {
			Start int `yaml:"start"`
			End   int `yaml:"end"`
		} `yaml:"portRange"`
	} `yaml:"pfSense"`

	Landing struct {
		Url      string `yaml:"url"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		ClientID string `yaml:"clientId"`
	} `yaml:"landing"`

	K8s struct {
		Config string `yaml:"config"`
	} `yaml:"k8s"`

	NPM struct {
		Url      string `yaml:"url"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"npm"`

	Harbor struct {
		Url           string `yaml:"url"`
		User          string `yaml:"user"`
		Password      string `yaml:"password"`
		WebhookSecret string `yaml:"webhookSecret"`
	} `yaml:"harbor"`
}

var Env Environment

func SetupEnvironment() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup environment. details: %s", err)
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

	err = yaml.Unmarshal(yamlFile, &Env)
	if err != nil {
		log.Fatalf(makeError(err).Error())
	}

	log.Println("config loaded")
}
