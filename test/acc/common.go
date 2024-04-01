package acc

import (
	"github.com/google/uuid"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"os"
	"strings"
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

func GenName(base ...string) string {
	if len(base) == 0 {
		return "acc-" + strings.ReplaceAll(uuid.NewString()[:10], "-", "")
	}

	return "acc-" + strings.ReplaceAll(base[0], " ", "-") + "-" + strings.ReplaceAll(uuid.NewString()[:10], "-", "")
}
