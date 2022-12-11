package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models"
	"go-deploy/pkg/app"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/deployment_worker"
	"go-deploy/routers"
	"log"
	"net/http"
	"os"
)

func setup(context *app.Context) {
	conf.Setup()
	models.Setup()
	k8s.Setup()
	deployment_worker.Setup(context)
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
