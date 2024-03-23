package app

import (
	"context"
	"errors"
	argFlag "flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/intializer"
	"go-deploy/pkg/metrics"
	"go-deploy/pkg/workers/migrate"
	"go-deploy/routers"
	"log"
	"net/http"
	"os"
	"time"
)

type Options struct {
	Flags    FlagDefinitionList
	TestMode bool
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
	Name      string
	Task      func() error
	Composite bool
}

func (it *InitTask) LogBegin(prefix string) {
	orange := "\033[38;5;208m"
	grey := "\033[90m"
	reset := "\033[0m"
	now := time.Now().Format("2006/01/02 15:04:05")
	taskName := it.Name
	fmt.Printf("[%s] %s %s%s%s %s...%s ", now, prefix, orange, taskName, reset, grey, reset)
	if it.Composite {
		lightBlue := "\033[38;5;39m"
		fmt.Println(lightBlue)
	}
}

func (it *InitTask) LogCompleted() {
	green := "\033[32m"
	grey := "\033[90m"
	reset := "\033[0m"
	if it.Composite {
		fmt.Printf("%s... done %s✓%s\n", grey, green, reset)
	} else {
		fmt.Printf("%s✓%s\n", green, reset)
	}
}

func (it *InitTask) LogFailed() {
	red := "\033[31m"
	grey := "\033[90m"
	reset := "\033[0m"
	if it.Composite {
		fmt.Printf("%s... failed %s✗%s\n", grey, red, reset)
	} else {
		fmt.Printf("%s✗%s\n", red, reset)
	}
}

// Create creates a new App instance.
func Create(opts *Options) *App {

	initTasks := []InitTask{
		{Name: "Setup environment", Task: config.SetupEnvironment, Composite: true},
		{Name: "Setup metrics", Task: metrics.Setup},
		{Name: "Enable test mode if requested", Task: enableTestIfRequested},
		{Name: "Setup database", Task: db.Setup, Composite: true},
		{Name: "Clean up old tests", Task: intializer.CleanUpOldTests},
		{Name: "Synchronize VM ports", Task: intializer.SynchronizeVmPorts, Composite: true},
		{Name: "Run migrations", Task: migrator.Migrate, Composite: true},
		{Name: "Reset running jobs", Task: func() error { return job_repo.New().ResetRunning() }},
	}

	for idx, task := range initTasks {
		task.LogBegin(fmt.Sprintf("(%d/%d)", idx+1, len(initTasks)))
		err := task.Task()
		if err != nil {
			task.LogFailed()
			log.Fatalln("Task", task.Name, "failed. See error below:\n", err)
		}
		task.LogCompleted()
	}

	ctx, cancel := context.WithCancel(context.Background())

	for _, flag := range opts.Flags {
		// handle api worker separately
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
			gin.SetMode("debug")
		}

		httpServer = &http.Server{
			Addr:    fmt.Sprintf("0.0.0.0:%d", config.Config.Port),
			Handler: routers.NewRouter(),
		}

		go func() {
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
			log.Println("waiting for http server to shutdown...")
		}
	}

	shutdown()

	log.Println("server exited successfully")
}

func ParseFlags() *Options {
	flags := GetFlags()

	for _, flag := range flags {
		switch flag.ValueType {
		case "bool":
			argFlag.Bool(flag.Name, flag.DefaultValue.(bool), flag.Description)
		}
	}
	argFlag.Parse()

	for _, flag := range flags {
		switch flag.ValueType {
		case "bool":
			if lookedUpVal := argFlag.Lookup(flag.Name); lookedUpVal != nil {
				flags.SetPassedValue(flag.Name, argFlag.Lookup(flag.Name).Value.(argFlag.Getter).Get().(bool))
			}
		}
	}

	options := Options{
		Flags:    flags,
		TestMode: flags.GetPassedValue("test-mode").(bool),
	}

	if options.TestMode {
		log.Println("RUNNING IN TEST MODE. NO AUTHENTICATION WILL BE REQUIRED.")
	}

	if !flags.AnyWorkerFlagsPassed() {
		log.Println("no workers specified, starting all")

		for _, flag := range flags {
			switch flag.FlagType {
			case "worker":
				flags.SetPassedValue(flag.Name, true)
			}
		}
	}

	return &options
}

func enableTestIfRequested() error {
	if config.Config.TestMode {
		config.Config.TestMode = true
		config.Config.MongoDB.Name = config.Config.MongoDB.Name + "-test"
	}

	return nil
}
