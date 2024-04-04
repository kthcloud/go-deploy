package cmd

import (
	"context"
	"errors"
	argFlag "flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/mode"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/migrate"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/intializer"
	"go-deploy/pkg/log"
	"go-deploy/pkg/metrics"
	"go-deploy/routers"
	"net/http"
	"os"
	"time"
)

type Options struct {
	Flags FlagDefinitionList
	Mode  string
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
		{Name: "Enable test mode if requested", Task: enableTestIfRequested},
		{Name: "Setup database", Task: db.Setup},
		{Name: "Clean up old tests", Task: intializer.CleanUpOldTests},
		{Name: "Synchronize VM ports", Task: intializer.SynchronizeVmPorts},
		{Name: "Run migrations", Task: migrator.Migrate},
		{Name: "Reset running jobs", Task: func() error { return job_repo.New().ResetRunning() }},
	}

	for idx, task := range initTasks {
		task.Begin(fmt.Sprintf("(%d/%d)", idx+1, len(initTasks)))
		err := task.Task()
		if err != nil {
			log.Fatalf("Init task %s failed. details: %s", task.Name, err.Error())
		}
	}
	log.Printf("%sInitialization completed%s", log.Orange, log.Reset)

	ctx, cancel := context.WithCancel(context.Background())

	for _, flag := range opts.Flags {
		// Handle api worker separately
		if flag.Name == "api" {
			continue
		}

		if flag.FlagType == "worker" && flag.GetPassedValue().(bool) {
			go flag.Run(ctx, cancel)
		}
	}

	var httpServer *http.Server

	if opts.Flags.GetPassedValue("api").(bool) {
		ginMode, exists := os.LookupEnv("GIN_MODE")
		if exists {
			gin.SetMode(ginMode)
		} else {
			gin.SetMode("release")
		}

		httpServer = &http.Server{
			Addr:    fmt.Sprintf("0.0.0.0:%d", config.Config.Port),
			Handler: routers.NewRouter(),
		}

		go func() {
			log.Printf("%sHTTP server listening on %s0.0.0.0:%d%s", log.Bold, log.Orange, config.Config.Port, log.Reset)
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := app.httpServer.Shutdown(ctx); err != nil {
			log.Fatalln(fmt.Errorf("failed to shutdown server. details: %w", err))
		}

		select {
		case <-ctx.Done():
			log.Println("Saiting for http server to shutdown...")
		}
	}

	shutdown()

	log.Println("Server exited successfully")
}

func ParseFlags() *Options {
	flags := GetFlags()

	// 1. Parse flags
	for _, flag := range flags {
		switch flag.ValueType {
		case "bool":
			argFlag.Bool(flag.Name, flag.DefaultValue.(bool), flag.Description)
		case "string":
			argFlag.String(flag.Name, flag.DefaultValue.(string), flag.Description)
		}
		// Add more cases as needed
	}
	argFlag.Parse()

	// 2. Extract passed values
	for _, flag := range flags {
		switch flag.ValueType {
		case "bool":
			if lookedUpVal := argFlag.Lookup(flag.Name); lookedUpVal != nil {
				flags.SetPassedValue(flag.Name, argFlag.Lookup(flag.Name).Value.(argFlag.Getter).Get().(bool))
			}
		case "string":
			if lookedUpVal := argFlag.Lookup(flag.Name); lookedUpVal != nil {
				flags.SetPassedValue(flag.Name, argFlag.Lookup(flag.Name).Value.(argFlag.Getter).Get().(string))
			}
		}
		// Add more cases as needed
	}

	options := Options{
		Flags: flags,
		Mode:  flags.GetPassedValue("mode").(string),
	}

	if options.Mode != mode.Test && options.Mode != mode.Prod && options.Mode != mode.Dev {
		panic("Invalid mode specified. Valid options are: test, dev, prod")
	}

	return &options
}

func validateApp(options *Options) error {
	if !options.Flags.AnyWorkerFlagsPassed() {
		log.Println("No workers specified, starting all")

		for _, flag := range options.Flags {
			switch flag.FlagType {
			case "worker":
				options.Flags.SetPassedValue(flag.Name, true)
			}
		}
	}

	return nil
}

func enableTestIfRequested() error {
	if config.Config.Mode == mode.Test {
		log.Printf("%sRUNNING IN TEST MODE. NO AUTHENTICATION WILL BE REQUIRED%s", log.Bold, log.Reset)
		config.Config.MongoDB.Name = config.Config.MongoDB.Name + "-test"
	}

	return nil
}
