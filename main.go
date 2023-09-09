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

	flag.Parse()

	api := isFlagPassed("api")
	confirmer := isFlagPassed("confirmer")
	statusUpdater := isFlagPassed("status-updater")
	jobExecutor := isFlagPassed("job-executor")
	repairer := isFlagPassed("repairer")
	pinger := isFlagPassed("pinger")
	snapshotter := isFlagPassed("snapshotter")

	var options *app.StartOptions
	if confirmer || statusUpdater || jobExecutor || repairer || api || pinger || snapshotter {
		options = &app.StartOptions{
			API:           api,
			Confirmer:     confirmer,
			StatusUpdater: statusUpdater,
			JobExecutor:   jobExecutor,
			Repairer:      repairer,
			Pinger:        pinger,
			Snapshotter:   snapshotter,
		}

		log.Println("api: ", options.API)
		log.Println("confirmer: ", options.Confirmer)
		log.Println("status-updater: ", options.StatusUpdater)
		log.Println("job-executor: ", options.JobExecutor)
		log.Println("repairer: ", options.Repairer)
		log.Println("pinger: ", options.Pinger)
		log.Println("snapshotter: ", options.Snapshotter)
	} else {
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
