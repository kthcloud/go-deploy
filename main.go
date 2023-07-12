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

	flag.Parse()

	api := isFlagPassed("api")
	confirmer := isFlagPassed("confirmer")
	statusUpdater := isFlagPassed("status-updater")
	jobExecutor := isFlagPassed("job-executor")
	repairer := isFlagPassed("repairer")

	var options *app.StartOptions
	if confirmer || statusUpdater || jobExecutor || repairer {
		options = &app.StartOptions{
			API:           api,
			Confirmer:     confirmer,
			StatusUpdater: statusUpdater,
			JobExecutor:   jobExecutor,
			Repairer:      repairer,
		}
	}

	log.Println("api: ", options.API)
	log.Println("confirmer: ", options.Confirmer)
	log.Println("status-updater: ", options.StatusUpdater)
	log.Println("job-executor: ", options.JobExecutor)
	log.Println("repairer: ", options.Repairer)

	server := app.Start(options)
	if server != nil {
		defer func() {
			app.Stop(server)
		}()
	}

	quit := make(chan os.Signal)
	<-quit
	log.Println("received shutdown signal")
}
