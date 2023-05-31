package repair

import (
	"go-deploy/pkg/app"
	"log"
)

func Setup(ctx *app.Context) {
	log.Println("starting repairer")
	go repairer(ctx)
}
