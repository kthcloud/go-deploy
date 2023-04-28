package main

import (
	"fmt"
	"go-deploy/models"
	"go-deploy/pkg/app"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/intializer"
	"go-deploy/pkg/workers/confirmers"
	"go-deploy/pkg/workers/status_updaters"
	"go-deploy/routers"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func setup(context *app.Context) {
	conf.SetupEnvironment()
	conf.SetupGPU()

	models.Setup()
	confirmers.Setup(context)
	status_updaters.Setup(context)

	intializer.SynchronizeGPU()
}

func shutdown() {
	models.Shutdown()
}

func main() {
	context := app.Context{}

	setup(&context)
	defer shutdown()

	ginMode, exists := os.LookupEnv("GIN_MODE")
	if exists {
		gin.SetMode(ginMode)
	} else {
		gin.SetMode("debug")
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", conf.Env.Port),
		Handler: routers.NewRouter(),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to start http server. details: %s\n", err)
	}

}
