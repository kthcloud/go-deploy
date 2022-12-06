package main

import (
	"deploy-api-go/models"
	"deploy-api-go/pkg/app"
	"deploy-api-go/pkg/conf"
	"deploy-api-go/pkg/subsystems/k8s"
	"deploy-api-go/pkg/worker"
	"deploy-api-go/routers"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func setup(context *app.Context) {
	conf.Setup()
	models.Setup()
	k8s.Setup()
	worker.Setup(context)
}

func shutdown() {
	models.Shutdown()
}

func main() {
	context := app.Context{}

	setup(&context)
	defer shutdown()

	gin.SetMode("debug")

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", conf.Env.Port),
		Handler: routers.NewRouter(),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to start http server. details: %s\n", err)
	}

}
