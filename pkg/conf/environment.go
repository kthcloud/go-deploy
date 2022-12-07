package conf

import (
	"fmt"
	env "github.com/Netflix/go-env"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Environment struct {
	Port        int    `env:"DEPLOY_PORT,default=8080"`
	ExternalUrl string `env:"DEPLOY_EXTERNAL_URL"`

	SessionSecret string `env:"DEPLOY_SESSION_SECRET,required=true"`
	ParentDomain  string `env:"DEPLOY_PARENT_DOMAIN,required=true"`

	DockerRegistry struct {
		Url              string `env:"DEPLOY_DOCKER_REGISTRY_URL,required=true"`
		PlaceHolderImage string `env:"DEPLOY_PLACEHOLDER_DOCKER_IMAGE,required=true"`
	}

	AppPort   int    `env:"DEPLOY_APP_PORT,default=8080"`
	AppPrefix string `env:"DEPLOY_APP_PREFIX,required=true"`

	Keycloak struct {
		Url   string `env:"DEPLOY_KEYCLOAK_URL,required=true"`
		Realm string `env:"DEPLOY_KEYCLOAK_REALM,required=true"`
	}

	K8s struct {
		Config string `env:"DEPLOY_K8S_CONFIG"`
	}

	Npm struct {
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

	Db struct {
		Url      string `env:"DEPLOY_DB_URL,required=true"`
		Name     string `env:"DEPLOY_DB_NAME,required=true"`
		Username string `env:"DEPLOY_DB_USERNAME"`
		Password string `env:"DEPLOY_DB_PASSWORD"`
	}
}

var Env Environment

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
}
