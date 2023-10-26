package metrics

import (
	"context"
	"go-deploy/models/sys/event"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/metrics"
	"log"
	"time"
)

func metricsWorker(ctx context.Context) {
	defer log.Println("metrics worker stopped")

	metricsFuncMap := map[string]func() error{
		"total_unique_users": setTotalUniqueUserMetrics,
	}

	for {
		select {
		case <-time.After(time.Duration(conf.Env.Metrics.Interval) * time.Second):
			for name, metric := range metricsFuncMap {
				if err := metric(); err != nil {
					log.Printf("error computing metric %s. details: %s", name, err.Error())
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func setTotalUniqueUserMetrics() error {
	// get distinct users by userId
	uniqueUsers, err := event.New().CountDistinct("source.userId")
	if err != nil {
		return err
	}

	metrics.UniqueUsers.Set(float64(uniqueUsers))

	return nil
}
