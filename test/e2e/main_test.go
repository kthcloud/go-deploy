package e2e

import (
	"go-deploy/pkg/app"
	"go-deploy/pkg/conf"
	"log"
	"os"
	"testing"
	"time"
)

var deployApp *app.App

func TestMain(m *testing.M) {
	Setup()
	code := m.Run()
	Shutdown()
	os.Exit(code)
}

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

	deployApp = app.Create(nil)
	if deployApp == nil {
		log.Fatalln("failed to create app")
	}

	// TODO: wait for server to start instead of using this "hack"
	time.Sleep(3 * time.Second)
}

func Shutdown() {
	if deployApp != nil {
		deployApp.Stop()
	}
}
