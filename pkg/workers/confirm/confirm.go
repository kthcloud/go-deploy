package confirm

import (
	"context"
	"log"
)

func Setup(ctx context.Context) {
	log.Println("starting confirmers")
	go deploymentConfirmer(ctx)
	go smConfirmer(ctx)
	go vmConfirmer(ctx)
}
