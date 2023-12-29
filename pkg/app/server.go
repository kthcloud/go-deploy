package app

import (
	"context"
	"errors"
	argFlag "flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/db"
	"go-deploy/models/sys/job"
	"go-deploy/pkg/config"
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

func Create(opts *Options) *App {
	config.SetupEnvironment()
	metrics.Setup()

	if opts.TestMode {
		config.Config.TestMode = true
		config.Config.MongoDB.Name = config.Config.MongoDB.Name + "-test"
	}

	db.Setup()
	intializer.CleanUpOldTests()
	migrator.Migrate()

	err := job.New().ResetRunning()
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to reset running job. details: %w", err))
	}

	intializer.SynchronizeGPUs()
	intializer.SynchronizeVmPorts()

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
			err = httpServer.ListenAndServe()
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
