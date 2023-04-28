package test

import (
	"go-deploy/pkg/conf"
	"os"
	"testing"
)

func setup(t *testing.T) {

	requiredEnvs := []string{
		"DEPLOY_CONFIG_FILE",
	}

	for _, env := range requiredEnvs {
		env, result := os.LookupEnv(env)
		if !result {
			t.Fatalf("%s must be set for acceptance test", env)
		}
	}

	_, result := os.LookupEnv("DEPLOY_CONFIG_FILE")
	if result {
		conf.SetupEnvironment()
	}
}
