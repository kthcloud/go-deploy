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

type Options struct {
	API           bool
	Confirmer     bool
	StatusUpdater bool
	JobExecutor   bool
	Repairer      bool
	Pinger        bool
	Snapshotter   bool

	TestMode bool
}

type App struct {
	httpServer *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
}

func shutdown() {
	models.Shutdown()
}

func Create(options *Options) *App {
	if options == nil || !options.TestMode {
		conf.SetupEnvironment()
	}

	models.Setup()

	migrator.Migrate()

	err := job.New().ResetRunning()
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to reset running job. details: %w", err))
	}

	intializer.SynchronizeGPUs()
	intializer.CleanUpOldTests()

	if options == nil {
		options = &Options{
			API:           true,
			Confirmer:     true,
			StatusUpdater: true,
			JobExecutor:   true,
			Repairer:      true,
			Pinger:        true,
			Snapshotter:   true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

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

		return &App{
			httpServer: server,
			ctx:        ctx,
			cancel:     cancel,
		}
	}

	return &App{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (app *App) Stop() {
	app.cancel()

	if app.httpServer != nil {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := app.httpServer.Shutdown(ctx); err != nil {
			log.Fatalln(fmt.Errorf("failed to shutdown server. details: %w", err))
		}

		select {
		case <-ctx.Done():
			log.Println("waiting for http server to shutdown...")
		}
	}

	shutdown()

	log.Println("server exited successfully")
}
