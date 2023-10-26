package metrics

import (
	"context"
	"log"
)

func Setup(ctx context.Context) {
	log.Println("starting metrics updaters")
	go metricsWorker(ctx)
}
