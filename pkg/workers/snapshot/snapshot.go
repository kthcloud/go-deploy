package snapshot

import (
	"go-deploy/pkg/sys"
	"log"
)

func Setup(ctx *sys.Context) {
	log.Println("starting snapshot workers")
	go snapshotter(ctx)
}
