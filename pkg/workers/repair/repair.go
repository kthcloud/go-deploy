package repair

import (
	"go-deploy/pkg/sys"
	"log"
)

func Setup(ctx *sys.Context) {
	log.Println("starting repairer")
	go deploymentRepairer(ctx)
	go vmRepairer(ctx)
	go gpuRepairer(ctx)
}
