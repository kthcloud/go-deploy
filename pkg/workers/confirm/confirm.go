package confirm

import (
	"go-deploy/pkg/app"
	"log"
)

func Setup(ctx *app.Context) {
	log.Println("starting confirm")
	go deploymentConfirmer(ctx)
	go vmConfirmer(ctx)
}
