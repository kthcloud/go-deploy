package main

import (
	"go-deploy/pkg/server"
	"log"
	"os"
)

func main() {
	httpServer := server.Start()

	quit := make(chan os.Signal)
	<-quit
	log.Println("received shutdown signal")

	server.Stop(httpServer)
}
