package main

import (
	"go-deploy/cmd"
	"go-deploy/pkg/log"
	"os"
)

func main() {
	options := cmd.ParseFlags()

	deployApp := cmd.Create(options)
	if deployApp == nil {
		log.Fatalln("failed to start app")
	}
	defer deployApp.Stop()

	quit := make(chan os.Signal)
	<-quit
	log.Println("received shutdown signal")
}
