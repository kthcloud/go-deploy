package acc

import (
	"go-deploy/pkg/config"
	"log"
	"os"
)

func Setup() {

	requiredEnvs := []string{
		"DEPLOY_CONFIG_FILE",
	}

	for _, env := range requiredEnvs {
		_, result := os.LookupEnv(env)
		if !result {
			log.Fatalln("required environment variable not set: " + env)
		}
	}

	_, result := os.LookupEnv("DEPLOY_CONFIG_FILE")
	if result {
		config.SetupEnvironment()
	}

	config.Config.TestMode = true
	config.Config.MongoDB.Name = config.Config.MongoDB.Name + "-test"
}

func Shutdown() {

}
