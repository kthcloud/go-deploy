package synchronize

import (
	"context"
	"log"
)

// Setup starts the synchronizers.
// Synchronizers are workers that periodically synchronize resources, such as GPUs.
func Setup(ctx context.Context) {
	log.Println("starting synchronizers")
	go gpuSynchronizer()
}
