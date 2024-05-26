package acc

import (
	"fmt"
	"go-deploy/models/mode"
	"go-deploy/pkg/config"
	"go-deploy/pkg/log"
	"os"
	"strings"

	"github.com/google/uuid"
)

func Setup() {

	requiredEnvs := []string{
		"DEPLOY_CONFIG_FILE",
	}

	for _, env := range requiredEnvs {
		_, result := os.LookupEnv(env)
		if !result {
			log.Fatalln("Required environment variable not set: " + env)
		}
	}

	err := log.SetupLogger(mode.Test)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup logger. details: %s", err.Error()))
	}

	if err := os.Chdir("../../../.."); err != nil {
		log.Fatalf("Failed to change working directory: %v", err)
	}

	_, result := os.LookupEnv("DEPLOY_CONFIG_FILE")
	if result {
		err := config.SetupEnvironment(mode.Test)
		if err != nil {
			log.Fatalln(err)
		}
	}

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
