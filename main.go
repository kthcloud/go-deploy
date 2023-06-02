package main

import (
	"go-deploy/pkg/app"
	"log"
	"os"
)

func main() {
	server := app.Start()
	defer func() {
		app.Stop(server)
	}()

	quit := make(chan os.Signal)
	<-quit
	log.Println("received shutdown signal")

}
