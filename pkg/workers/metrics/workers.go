package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go-deploy/models/sys/event"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/metrics"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func metricsWorker(ctx context.Context) {
	defer log.Println("metrics worker stopped")

	metricsFuncMap := map[string]func() error{
		"total_unique_users":   setTotalUniqueUserMetrics,
		"unique_users_by_date": setUniqueUsersByDate,
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

func setUniqueUsersByDate() error {
	type uniqueUserByDate struct {
		Filter bson.D
		Metric prometheus.Gauge
	}

	today := time.Now()

	uniqueUsersByDate := []uniqueUserByDate{
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())},
			}}},
			Metric: metrics.UniqueUsersToday,
		},
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Now().AddDate(0, 0, -1)},
			}}},
			Metric: metrics.UniqueUsersYesterday,
		},
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Now().AddDate(0, 0, -7)},
			}}},
			Metric: metrics.UniqueUsersThisWeek,
		},
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Now().AddDate(0, -1, 0)},
			}}},
			Metric: metrics.UniqueUsersThisMonth,
		},
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Now().AddDate(-1, 0, 0)},
			}}},
			Metric: metrics.UniqueUsersThisYear,
		},
	}

	for _, uniqueUser := range uniqueUsersByDate {
		count, err := event.New().AddExtraFilter(uniqueUser.Filter).CountDistinct("source.userId")
		if err != nil {
			return err
		}

		uniqueUser.Metric.Set(float64(count))
	}

	return nil
}
