package main

import (
	"go-deploy/cmd"
	"os"
)

func main() {
	options := cmd.ParseFlags()

	deployApp := cmd.Create(options)
	if deployApp == nil {
		panic("failed to start app")
	}
	defer deployApp.Stop()

	quit := make(chan os.Signal)
	<-quit
}
