package e2e

import (
	"go-deploy/pkg/log"
	"os"
)

func Setup() {
	//goland:noinspection ALL
	requiredEnvs := []string{}

	for _, env := range requiredEnvs {
		_, result := os.LookupEnv(env)
		if !result {
			log.Fatalln("required environment variable not set: " + env)
		}
	}
}

func Shutdown() {
}
