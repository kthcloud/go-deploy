package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models"
	"go-deploy/models/sys/job"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/intializer"
	"go-deploy/pkg/workers/confirm"
	"go-deploy/pkg/workers/job_execute"
	"go-deploy/pkg/workers/migrate"
	"go-deploy/pkg/workers/ping"
	"go-deploy/pkg/workers/repair"
	"go-deploy/pkg/workers/snapshot"
	"go-deploy/pkg/workers/status_update"
	"go-deploy/routers"
	"log"
	"net/http"
	"os"
	"time"
)

type StartOptions struct {
	API           bool
	Confirmer     bool
	StatusUpdater bool
	JobExecutor   bool
	Repairer      bool
	Pinger        bool
	Snapshotter   bool
}

func shutdown() {
	models.Shutdown()
}

func Start(ctx context.Context, options *StartOptions) *http.Server {
	conf.SetupEnvironment()

	models.Setup()

	migrator.Migrate()

	err := job.ResetRunning()
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to reset running job. details: %w", err))
	}

	intializer.SynchronizeGPUs()
	intializer.CleanUpOldTests()

	if options == nil {
		options = &StartOptions{
			API:           true,
			Confirmer:     true,
			StatusUpdater: true,
			JobExecutor:   true,
			Repairer:      true,
			Pinger:        true,
			Snapshotter:   true,
		}
	}

	if options.Confirmer {
		confirm.Setup(ctx)
	}
	if options.StatusUpdater {
		status_update.Setup(ctx)
	}
	if options.JobExecutor {
		job_execute.Setup(ctx)
	}
	if options.Repairer {
		repair.Setup(ctx)
	}
	if options.Pinger {
		ping.Setup(ctx)
	}
	if options.Snapshotter {
		snapshot.Setup(ctx)
	}
	if options.API {
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
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalln(fmt.Errorf("failed to start http server. details: %w", err))
			}
		}()

		return server
	}

	return nil
}

func Stop(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown server. details: %w\n", err)
	}

	select {
	case <-ctx.Done():
		log.Println("waiting for server to shutdown...")
	}

	shutdown()

	log.Println("server exited successfully")
}
