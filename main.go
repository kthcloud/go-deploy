package main

import (
	"go-deploy/pkg/app"
	"log"
	"os"
)

func main() {
	options := app.ParseFlags()

	deployApp := app.Create(options)
	if deployApp == nil {
		log.Fatalln("failed to start app")
	}
	defer deployApp.Stop()

	quit := make(chan os.Signal)
	<-quit
	log.Println("received shutdown signal")
}
