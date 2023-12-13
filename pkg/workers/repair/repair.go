package repair

import (
	"context"
	"log"
)

func Setup(ctx context.Context) {
	log.Println("starting repairers")
	go deploymentRepairer(ctx)
	go smRepairer(ctx)
	go vmRepairer(ctx)
}
