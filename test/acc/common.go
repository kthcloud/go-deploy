package acc

import (
	"go-deploy/pkg/conf"
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
		conf.SetupEnvironment()
	}

	conf.Env.TestMode = true
	conf.Env.DB.Name = conf.Env.DB.Name + "-test"
}

func Shutdown() {

}
