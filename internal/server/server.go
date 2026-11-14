package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

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
	"go.uber.org/zap"
)

// Options configures which server components and background workers are enabled.
type Options struct {
	Logger            *zap.Logger
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

// App manages the lifetime of the HTTP server and background workers.
type App struct {
	logger     *zap.Logger
	httpServer *http.Server
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	errCh      chan error
	shutdownFn func(context.Context) error
	stopOnce   sync.Once
	stopErr    error
}

type initTask struct {
	name string
	run  func(context.Context, *zap.Logger) error
}

type worker struct {
	enabled bool
	run     func(context.Context)
}

// Start initializes dependencies, starts enabled workers and the HTTP server,
// and binds their lifetimes to ctx.
func Start(ctx context.Context, opts Options) (*App, error) {
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	if err := log.SetupLogger(opts.Mode); err != nil {
		return nil, fmt.Errorf("setup logger: %w", err)
	}

	if err := validateOptions(&opts, opts.Logger); err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}

	tasks := []initTask{
		{name: "Setup environment", run: func(context.Context, *zap.Logger) error { return config.SetupEnvironment(opts.Mode) }},
		{name: "Setup metrics", run: func(context.Context, *zap.Logger) error { return metrics.Setup() }},
		{name: "Setup database", run: db.Setup},
		{name: "Clean up old tests", run: func(context.Context, *zap.Logger) error { return intializer.CleanUpOldTests() }},
		{name: "Synchronize VM ports", run: func(context.Context, *zap.Logger) error { return intializer.SynchronizeVmPorts() }},
		{name: "Run migrations", run: migrator.Migrate},
		{name: "Reset running jobs", run: func(context.Context, *zap.Logger) error { return job_repo.New().ResetRunning() }},
		{name: "Ensure system deployments exists", run: func(context.Context, *zap.Logger) error { return intializer.EnsureSystemDeploymentsExists() }},
		{name: "Ensure test users exist", run: func(context.Context, *zap.Logger) error { return intializer.EnsureTestUsersExist() }},
	}

	workers := []worker{
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

	var newHTTPServer func() *http.Server
	if opts.API {
		newHTTPServer = func() *http.Server {
			if ginMode, exists := os.LookupEnv("GIN_MODE"); exists {
				gin.SetMode(ginMode)
			} else {
				gin.SetMode(defaults.DefaultGinMode)
			}
			listenAddr := fmt.Sprintf("%s:%d", defaults.DefaultListenHost, config.Config.Port)
			return &http.Server{
				Addr:    listenAddr,
				Handler: routers.NewRouter(),
			}
		}
	}

	return startWith(ctx, opts.Logger, tasks, workers, newHTTPServer, db.Shutdown)
}

func startWith(
	ctx context.Context,
	l *zap.Logger,
	tasks []initTask,
	workers []worker,
	newHTTPServer func() *http.Server,
	shutdownFn func(context.Context) error,
) (*App, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if l == nil {
		l = zap.NewNop()
	}

	for idx, task := range tasks {
		l.Info("running init task", zap.Int("step", idx+1), zap.Int("total", len(tasks)), zap.String("task", task.name))
		if err := task.run(ctx, l); err != nil {
			return nil, fmt.Errorf("init task %q: %w", task.name, err)
		}
	}
	l.Info("initialization completed")

	var httpServer *http.Server
	if newHTTPServer != nil {
		httpServer = newHTTPServer()
	}

	appCtx, cancel := context.WithCancel(ctx)
	app := &App{
		logger:     l,
		httpServer: httpServer,
		cancel:     cancel,
		errCh:      make(chan error, 1),
		shutdownFn: shutdownFn,
	}

	for _, w := range workers {
		if !w.enabled || w.run == nil {
			continue
		}
		app.wg.Add(1)
		go func(run func(context.Context)) {
			defer app.wg.Done()
			run(appCtx)
		}(w.run)
	}

	if app.httpServer != nil {
		app.httpServer.BaseContext = func(_ net.Listener) context.Context {
			return appCtx
		}
		app.wg.Add(1)
		go func() {
			defer app.wg.Done()
			l.Info("HTTP server listening", zap.String("addr", app.httpServer.Addr))
			if err := app.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				l.Error("HTTP server failed", zap.Error(err))
				select {
				case app.errCh <- fmt.Errorf("http server: %w", err):
				default:
				}
				cancel()
			}
		}()
	}

	return app, nil
}

// Errors returns a channel that emits fatal asynchronous server errors (such as HTTP listen failures).
func (app *App) Errors() <-chan error {
	return app.errCh
}

// Stop cancels the application context, gracefully shuts down the HTTP server,
// waits for all background workers to exit, and runs resource cleanup.
func (app *App) Stop() error {
	app.stopOnce.Do(func() {
		app.cancel()

		var errs []error
		if app.httpServer != nil {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), defaults.HTTPShutdownTimeout)
			defer shutdownCancel()

			if err := app.httpServer.Shutdown(shutdownCtx); err != nil {
				errs = append(errs, fmt.Errorf("shutdown http server: %w", err))
				if closeErr := app.httpServer.Close(); closeErr != nil {
					errs = append(errs, fmt.Errorf("force close http server: %w", closeErr))
				}
			}
		}

		app.wg.Wait()

		if app.shutdownFn != nil {
			dbCtx, dbCancel := context.WithTimeout(context.Background(), defaults.HTTPShutdownTimeout)
			defer dbCancel()
			if err := app.shutdownFn(dbCtx); err != nil {
				errs = append(errs, err)
			}
		}

		app.logger.Info("server exited")
		app.stopErr = errors.Join(errs...)
	})
	return app.stopErr
}

func validateOptions(options *Options, l *zap.Logger) error {
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
		l.Info("no workers specified, starting all")

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
