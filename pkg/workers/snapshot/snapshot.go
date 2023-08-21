package snapshot

import (
	"context"
	"log"
)

func Setup(ctx context.Context) {
	log.Println("starting snapshot workers")
	go snapshotter(ctx)
}
