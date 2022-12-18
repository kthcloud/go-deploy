package conf

import (
	"fmt"
	env "github.com/Netflix/go-env"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
)

type Environment struct {
	Port        int    `env:"DEPLOY_PORT,default=8080"`
	ExternalUrl string `env:"DEPLOY_EXTERNAL_URL"`

	SessionSecret string `env:"DEPLOY_SESSION_SECRET,required=true"`
	ParentDomain  string `env:"DEPLOY_PARENT_DOMAIN,required=true"`

	DockerRegistry struct {
		Url                   string `env:"DEPLOY_DOCKER_REGISTRY_URL,required=true"`
		PlaceHolder           string `env:"DEPLOY_PLACEHOLDER_DOCKER_IMAGE,required=true"`
		PlaceHolderProject    string
		PlaceHolderRepository string
	}

	AppPort   int    `env:"DEPLOY_APP_PORT,default=8080"`
	AppPrefix string `env:"DEPLOY_APP_PREFIX,required=true"`

	Keycloak struct {
		Url   string `env:"DEPLOY_KEYCLOAK_URL,required=true"`
		Realm string `env:"DEPLOY_KEYCLOAK_REALM,required=true"`
	}

	Terraform struct {
		Url      string `env:"DEPLOY_TERRAFORM_DB_URL"`
		Username string `env:"DEPLOY_TERRAFORM_DB_USERNAME"`
		Password string `env:"DEPLOY_TERRAFORM_DB_PASSWORD"`
	}

	K8s struct {
		Config string `env:"DEPLOY_K8S_CONFIG"`
	}

	NPM struct {
		Identity string `env:"DEPLOY_NPM_ADMIN_IDENTITY,required=true"`
		Secret   string `env:"DEPLOY_NPM_ADMIN_SECRET,required=true"`
		Url      string `env:"DEPLOY_NPM_API_URL,required=true"`
	}

	Harbor struct {
		Identity      string `env:"DEPLOY_HARBOR_ADMIN_IDENTITY,required=true"`
		Secret        string `env:"DEPLOY_HARBOR_ADMIN_SECRET,required=true"`
		Url           string `env:"DEPLOY_HARBOR_API_URL,required=true"`
		WebhookSecret string `env:"DEPLOY_HARBOR_WEBHOOK_SECRET,required=true"`
	}

	PfSense struct {
		Identity       string `env:"DEPLOY_PFSENSE_ADMIN_IDENTITY,required=true"`
		Secret         string `env:"DEPLOY_PFSENSE_ADMIN_SECRET,required=true"`
		Url            string `env:"DEPLOY_PFSENSE_API_URL,required=true"`
		PublicIP       string `env:"DEPLOY_PFSENSE_PUBLIC_IP,required=true"`
		PortRange      string `env:"DEPLOY_PFSENSE_PORT_RANGE,required=true"`
		PortRangeStart int
		PortRangeEnd   int
	}

	CS struct {
		Url    string `env:"DEPLOY_CS_API_URL,required=true"`
		Key    string `env:"DEPLOY_CS_API_KEY,required=true"`
		Secret string `env:"DEPLOY_CS_SECRET_KEY,required=true"`
		ZoneID string `env:"DEPLOY_CS_ZONE_ID"`
	}

	DB struct {
		Url      string `env:"DEPLOY_DB_URL,required=true"`
		Name     string `env:"DEPLOY_DB_NAME,required=true"`
		Username string `env:"DEPLOY_DB_USERNAME"`
		Password string `env:"DEPLOY_DB_PASSWORD"`
	}
}

var Env Environment

func dockerRegistrySetup() {
	pfsenseRangeError := "docker registry placeholder image must be specified as project:repository]"

	placeholderImageSplit := strings.Split(Env.DockerRegistry.PlaceHolder, ":")

	if len(placeholderImageSplit) != 2 {
		log.Fatalln(pfsenseRangeError)
	}

	Env.DockerRegistry.PlaceHolderProject = placeholderImageSplit[0]
	Env.DockerRegistry.PlaceHolderRepository = placeholderImageSplit[1]
}

func pfsenseSetup() {
	portRangeSplit := strings.Split(Env.PfSense.PortRange, "-")

	pfsenseRangeError := "pfsense port range must be specified as (start-end]"

	if len(portRangeSplit) != 2 {
		log.Fatalln(pfsenseRangeError)
	}

	start, err := strconv.Atoi(portRangeSplit[0])
	if err != nil {
		log.Fatalln(pfsenseRangeError)
	}

	end, err := strconv.Atoi(portRangeSplit[1])
	if err != nil {
		log.Fatalln(pfsenseRangeError)
	}

	Env.PfSense.PortRangeStart = start
	Env.PfSense.PortRangeEnd = end
}

func Setup() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup environment. details: %s", err)
	}

	deployEnv, found := os.LookupEnv("DEPLOY_ENV_FILE")
	if found {
		log.Println("using env-file:", deployEnv)
		err := godotenv.Load(deployEnv)
		if err != nil {
			log.Fatalln(makeError(err))
		}
	}

	_, err := env.UnmarshalFromEnviron(&Env)
	if err != nil {
		log.Fatalln(makeError(err))
	}

	dockerRegistrySetup()
	pfsenseSetup()
}
