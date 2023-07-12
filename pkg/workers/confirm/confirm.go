package confirm

import (
	"go-deploy/pkg/sys"
	"log"
)

func Setup(ctx *sys.Context) {
	log.Println("starting confirmers")
	go deploymentConfirmer(ctx)
	go vmConfirmer(ctx)
}
