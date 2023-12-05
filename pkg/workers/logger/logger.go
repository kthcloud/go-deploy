package logger

import (
	"context"
	"log"
)

func Setup(ctx context.Context) {
	log.Println("starting job workers")
	go deploymentLogger(ctx)
}
