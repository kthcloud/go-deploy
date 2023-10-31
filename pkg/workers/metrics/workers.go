package metrics

import (
	"context"
	"fmt"
	"go-deploy/models/sys/event"
	"go-deploy/models/sys/key_value"
	"go-deploy/pkg/config"
	"go-deploy/pkg/metrics"
	"go-deploy/utils"
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
		case <-time.After(time.Duration(config.Config.Metrics.Interval) * time.Second):
			for name, metric := range metricsFuncMap {
				if err := metric(); err != nil {
					utils.PrettyPrintError(fmt.Errorf("error computing metric %s. details: %w", name, err))
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
		Key    string
	}

	today := time.Now()

	uniqueUsersByDate := []uniqueUserByDate{
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())},
			}}},
			Key: "metrics:unique-users:today",
		},
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Now().AddDate(0, 0, -1)},
			}}},
			Key: "metrics:unique-users:yesterday",
		},
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Now().AddDate(0, 0, -7)},
			}}},
			Key: "metrics:unique-users:this-week",
		},
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Now().AddDate(0, -1, 0)},
			}}},
			Key: "metrics:unique-users:this-month",
		},
		{
			Filter: bson.D{{"createdAt", bson.D{
				{"$gte", time.Now().AddDate(-1, 0, 0)},
			}}},
			Key: "metrics:unique-users:this-year",
		},
	}

	for _, uniqueUser := range uniqueUsersByDate {
		count, err := event.New().AddExtraFilter(uniqueUser.Filter).CountDistinct("source.userId")
		if err != nil {
			return err
		}

		err = key_value.New().Set(uniqueUser.Key, count, 0)
		if err != nil {
			return fmt.Errorf("failed to set key %s when fetching metrics. details: %w", uniqueUser.Key, err)
		}
	}

	return nil
}
