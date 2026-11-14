package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/internal/defaults"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db"
	migrator "github.com/kthcloud/go-deploy/pkg/db/migrate"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/pkg/intializer"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/metrics"
	"github.com/kthcloud/go-deploy/pkg/services/cleaner"
	"github.com/kthcloud/go-deploy/pkg/services/confirm"
	"github.com/kthcloud/go-deploy/pkg/services/job_execute"
	"github.com/kthcloud/go-deploy/pkg/services/job_schedule"
	"github.com/kthcloud/go-deploy/pkg/services/logger"
	metricsWorker "github.com/kthcloud/go-deploy/pkg/services/metrics_update"
	"github.com/kthcloud/go-deploy/pkg/services/status_update"
	"github.com/kthcloud/go-deploy/pkg/services/synchronize"
	"github.com/kthcloud/go-deploy/pkg/services/system_state_poll"
	"github.com/kthcloud/go-deploy/routers"
	"github.com/kthcloud/go-deploy/service/utils"
)

type Options struct {
	Ctx               context.Context
	Mode              string
	API               bool
	Confirmer         bool
	StatusUpdater     bool
	Synchronizer      bool
	JobExecutor       bool
	Repairer          bool
	MetricsUpdater    bool
	LoggerControl     bool
	LoggerWorker      bool
	Cleaner           bool
	SystemStatePoller bool
}

type App struct {
	httpServer *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
}

func shutdown() {
	db.Shutdown()
}

type InitTask struct {
	Name string
	Task func() error
}

func (it *InitTask) Begin(prefix string) {
	log.Infof("%s %s%s%s %s...%s ", prefix, log.Orange, it.Name, log.Reset, log.Grey, log.Reset)
}

// Create creates a new App instance.
func Create(opts *Options) *App {
	err := log.SetupLogger(opts.Mode)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup logger. details: %s", err.Error()))
	}

	initTasks := []InitTask{
		{Name: "Validate application", Task: func() error { return validateApp(opts) }},
		{Name: "Setup environment", Task: func() error { return config.SetupEnvironment(opts.Mode) }},
		{Name: "Setup metrics", Task: metrics.Setup},
		{Name: "Setup database", Task: db.Setup},
		{Name: "Clean up old tests", Task: intializer.CleanUpOldTests},
		{Name: "Synchronize VM ports", Task: intializer.SynchronizeVmPorts},
		{Name: "Run migrations", Task: migrator.Migrate},
		{Name: "Reset running jobs", Task: func() error { return job_repo.New().ResetRunning() }},
		{Name: "Ensure system deployments exists", Task: intializer.EnsureSystemDeploymentsExists},
		{Name: "Ensure test users exist", Task: intializer.EnsureTestUsersExist},
	}

	for idx, task := range initTasks {
		task.Begin(fmt.Sprintf("(%d/%d)", idx+1, len(initTasks)))
		err := task.Task()
		if err != nil {
			log.Fatalf("Init task %s failed. details: %s", task.Name, err.Error())
		}
	}
	log.Printf("%sInitialization completed%s", log.Orange, log.Reset)

	ctx, cancel := context.WithCancel(utils.FirstNonZero(opts.Ctx, context.Background()))
	workers := []struct {
		enabled bool
		run     func(context.Context)
	}{
		{opts.Confirmer, confirm.Setup},
		{opts.StatusUpdater, status_update.Setup},
		{opts.Synchronizer, synchronize.Setup},
		{opts.JobExecutor, job_execute.Setup},
		{opts.Repairer, job_schedule.Setup},
		{opts.MetricsUpdater, metricsWorker.Setup},
		{opts.LoggerControl, func(ctx context.Context) { logger.Setup(ctx, []logger.LogRole{logger.LogRoleControl}) }},
		{opts.LoggerWorker, func(ctx context.Context) { logger.Setup(ctx, []logger.LogRole{logger.LogRoleWorker}) }},
		{opts.Cleaner, cleaner.Setup},
		{opts.SystemStatePoller, system_state_poll.Setup},
	}
	for _, w := range workers {
		if w.enabled {
			go w.run(ctx)
		}
	}

	var httpServer *http.Server

	if opts.API {
		ginMode, exists := os.LookupEnv("GIN_MODE")
		if exists {
			gin.SetMode(ginMode)
		} else {
			gin.SetMode(defaults.DefaultGinMode)
		}

		listenAddr := fmt.Sprintf("%s:%d", defaults.DefaultListenHost, config.Config.Port)
		httpServer = &http.Server{
			Addr:    listenAddr,
			Handler: routers.NewRouter(),
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
		}

		go func() {
			log.Printf("%sHTTP server listening on %s%s%s", log.Bold, log.Orange, listenAddr, log.Reset)
			err := httpServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalln(fmt.Errorf("failed to start http server. details: %w", err))
			}
		}()
	}

	return &App{
		httpServer: httpServer,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (app *App) Stop() {
	app.cancel()

	if app.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaults.HTTPShutdownTimeout)
		defer cancel()
		if err := app.httpServer.Shutdown(ctx); err != nil {
			log.Errorln(fmt.Errorf("failed to gracefully shutdown server. details: %w", err))
			if closeErr := app.httpServer.Close(); closeErr != nil {
				log.Fatalln("Force close failed:", closeErr)
			}
		}

		<-ctx.Done()
		log.Println("Waiting for http server to shutdown...")
	}

	shutdown()

	log.Println("Server exited successfully")
}

func validateApp(options *Options) error {
	anyWorker := options.API ||
		options.Confirmer ||
		options.StatusUpdater ||
		options.Synchronizer ||
		options.JobExecutor ||
		options.Repairer ||
		options.MetricsUpdater ||
		options.LoggerControl ||
		options.LoggerWorker ||
		options.Cleaner ||
		options.SystemStatePoller

	if !anyWorker {
		log.Println("No workers specified, starting all")

		options.API = true
		options.Confirmer = true
		options.StatusUpdater = true
		options.Synchronizer = true
		options.JobExecutor = true
		options.Repairer = true
		options.MetricsUpdater = true
		options.LoggerControl = true
		options.LoggerWorker = true
		options.Cleaner = true
		options.SystemStatePoller = true
	}

	return nil
}
