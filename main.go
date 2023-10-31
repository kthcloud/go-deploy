package main

import (
	"flag"
	"go-deploy/pkg/app"
	"log"
	"os"
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	_ = flag.Bool("api", false, "start api")
	_ = flag.Bool("confirmer", false, "start confirmer")
	_ = flag.Bool("status-updater", false, "start status updater")
	_ = flag.Bool("job-executor", false, "start job executor")
	_ = flag.Bool("repairer", false, "start repairer")
	_ = flag.Bool("pinger", false, "start pinger")
	_ = flag.Bool("snapshotter", false, "start snapshotter")
	_ = flag.Bool("metrics-updater", false, "start metrics updater")
	_ = flag.Bool("test-mode", false, "run in test mode")

	flag.Parse()

	api := isFlagPassed("api")
	confirmer := isFlagPassed("confirmer")
	statusUpdater := isFlagPassed("status-updater")
	jobExecutor := isFlagPassed("job-executor")
	repairer := isFlagPassed("repairer")
	pinger := isFlagPassed("pinger")
	snapshotter := isFlagPassed("snapshotter")
	metricsUpdater := isFlagPassed("metrics-updater")
	testMode := isFlagPassed("test-mode")

	options := app.Options{
		Workers:  app.Workers{},
		TestMode: testMode,
	}

	if testMode {
		log.Println("RUNNING IN TEST MODE. NO AUTHENTICATION WILL BE REQUIRED.")
	}

	if confirmer || statusUpdater || jobExecutor || repairer || api || pinger || snapshotter || metricsUpdater {
		options.Workers = app.Workers{
			API:            api,
			Confirmer:      confirmer,
			StatusUpdater:  statusUpdater,
			JobExecutor:    jobExecutor,
			Repairer:       repairer,
			Pinger:         pinger,
			Snapshotter:    snapshotter,
			MetricsUpdater: metricsUpdater,
		}

		workers := &options.Workers

		log.Println("api: ", workers.API)
		log.Println("confirmer: ", workers.Confirmer)
		log.Println("status-updater: ", workers.StatusUpdater)
		log.Println("job-executor: ", workers.JobExecutor)
		log.Println("repairer: ", workers.Repairer)
		log.Println("pinger: ", workers.Pinger)
		log.Println("snapshotter: ", workers.Snapshotter)
		log.Println("metrics-updater: ", workers.MetricsUpdater)

	} else {
		options.Workers = app.Workers{
			API:            true,
			Confirmer:      true,
			StatusUpdater:  true,
			JobExecutor:    true,
			Repairer:       true,
			Pinger:         true,
			Snapshotter:    true,
			MetricsUpdater: true,
		}

		log.Println("no workers specified, starting all")
	}

	deployApp := app.Create(options)
	if deployApp == nil {
		log.Fatalln("failed to start app")
	}
	defer deployApp.Stop()

	quit := make(chan os.Signal)
	<-quit
	log.Println("received shutdown signal")
}
