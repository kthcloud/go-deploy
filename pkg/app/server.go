package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models"
	"go-deploy/models/sys/job"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/intializer"
	"go-deploy/pkg/sys"
	"go-deploy/pkg/workers/confirm"
	"go-deploy/pkg/workers/job_execute"
	"go-deploy/pkg/workers/repair"
	"go-deploy/pkg/workers/status_update"
	"go-deploy/routers"
	"log"
	"net/http"
	"os"
	"time"
)

func setup(context *sys.Context) {
	conf.SetupEnvironment()

	models.Setup()
	err := job.ResetRunning()
	if err != nil {
		log.Fatalln("failed to reset running job. details: ", err)
	}

	intializer.SynchronizeGPUs()

	confirm.Setup(context)
	status_update.Setup(context)
	job_execute.Setup(context)
	repair.Setup(context)
}

func shutdown() {
	models.Shutdown()
}

func Start() *http.Server {
	appContext := sys.Context{}

	setup(&appContext)

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

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start http server. details: %s\n", err)
		}
	}()

	return server
}

func Stop(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown server. details: %s\n", err)
	}

	select {
	case <-ctx.Done():
		log.Println("waiting for server to shutdown...")
	}

	shutdown()

	log.Println("server exited successfully")
}
