package metrics

import (
	"context"
	"fmt"
	"go-deploy/models/sys/event"
	"go-deploy/models/sys/job"
	"go-deploy/models/sys/key_value"
	"go-deploy/pkg/config"
	"go-deploy/pkg/metrics"
	"go-deploy/pkg/workers"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

// metricsWorker is a worker that updates metrics.
func metricsWorker(ctx context.Context) {
	defer workers.OnStop("metrics worker")

	metricsFuncMap := map[string]func() error{
		"users-total":          usersTotal,
		"daily-active-users":   activeUsers("daily", metrics.KeyDailyActiveUsers, 2),
		"monthly-active-users": activeUsers("monthly", metrics.KeyMonthlyActiveUsers, 2),
		"jobs-total":           jobMetrics(metrics.KeyJobsTotal, nil),
		"jobs-pending":         jobMetrics(metrics.KeyJobsPending, strPtr(job.StatusPending)),
		"jobs-running":         jobMetrics(metrics.KeyJobsRunning, strPtr(job.StatusRunning)),
		"jobs-failed":          jobMetrics(metrics.KeyJobsFailed, strPtr(job.StatusFailed)),
		"jobs-terminated":      jobMetrics(metrics.KeyJobsTerminated, strPtr(job.StatusTerminated)),
		"jobs-completed":       jobMetrics(metrics.KeyJobsCompleted, strPtr(job.StatusCompleted)),
	}

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(time.Duration(config.Config.Metrics.Interval) * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("metricsWorker")

		case <-tick:
			for metricGroupName, metric := range metricsFuncMap {
				if err := metric(); err != nil {
					utils.PrettyPrintError(fmt.Errorf("error computing metric %s. details: %w", metricGroupName, err))
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// usersTotal computes the total number of users and stores it in the key-value store.
func usersTotal() error {
	total, err := event.New().CountDistinct("source.userId")
	if err != nil {
		return fmt.Errorf("error counting distinct users when computing metrics. details: %w", err)
	}

	err = key_value.New().Set(metrics.KeyUsersTotal, fmt.Sprintf("%d", total), 0)
	if err != nil {
		return fmt.Errorf("error setting value for key %s when computing metrics. details: %w", metrics.KeyUsersTotal, err)
	}

	return nil
}

// activeUsers computes the number of active users and stores it in the key-value store.
func activeUsers(frequencyType, key string, count int) func() error {
	return func() error {
		pipeline := getActiveUserMongoPipeline(frequencyType, count)

		cursor, err := event.New().Collection.Aggregate(context.Background(), pipeline)
		if err != nil {
			log.Fatal(err)
		}
		defer func(cursor *mongo.Cursor, ctx context.Context) {
			err = cursor.Close(ctx)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error closing cursor when fetching metrics. details: %w", err))
			}
		}(cursor, context.Background())

		// Iterate over the results
		for cursor.Next(context.Background()) {
			var result bson.M
			if err := cursor.Decode(&result); err != nil {
				log.Fatal(err)
			}

			total := result["total"].(int32)
			err = key_value.New().Set(key, fmt.Sprintf("%d", total), 0)
			if err != nil {
				return fmt.Errorf("error setting value for key %s when computing metrics. details: %w", key, err)
			}
		}

		return nil
	}
}

// getActiveUserMongoPipeline returns a mongo pipeline that computes the number of active users.
func getActiveUserMongoPipeline(frequencyType string, count int) mongo.Pipeline {
	var gte time.Time
	var lte time.Time
	var dateFormat string

	now := time.Now()

	switch frequencyType {
	case "daily":
		gte = time.Now().AddDate(0, 0, -count)
		gte = time.Date(gte.Year(), gte.Month(), gte.Day(), 0, 0, 0, 0, time.UTC)
		lte = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		dateFormat = "%Y-%m-%d"
	case "monthly":
		gte = time.Now().AddDate(0, -count, 0)
		gte = time.Date(gte.Year(), gte.Month(), 1, 0, 0, 0, 0, time.UTC)
		lte = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		dateFormat = "%Y-%m"
	}

	return mongo.Pipeline{
		bson.D{
			{"$match", bson.M{
				"$and": bson.A{
					bson.M{"createdAt": bson.M{"$gte": gte}},
					bson.M{"createdAt": bson.M{"$lte": lte}},
				},
			}},
		},
		bson.D{
			{"$group", bson.M{
				"_id": bson.M{
					"userId": "$source.userId",
					"day": bson.M{
						"$dateToString": bson.M{
							"format": dateFormat,
							"date":   "$createdAt",
						},
					},
				},
			}},
		},
		bson.D{
			{"$group", bson.M{
				"_id": "$_id.userId",
				"count": bson.M{
					"$sum": 1,
				},
			}},
		},
		bson.D{
			{"$match", bson.M{
				"count": count,
			}},
		},
		bson.D{
			{"$count", "total"},
		},
	}
}

// jobMetrics computes the number of jobs and stores it in the key-value store.
func jobMetrics(key string, status *string) func() error {
	return func() error {
		filter := bson.D{}
		if status != nil {
			filter = append(filter, bson.E{Key: "status", Value: *status})
		}

		total, err := job.New().AddFilter(filter).Count()
		if err != nil {
			if status == nil {
				return fmt.Errorf("error counting jobs when computing metrics. details: %w", err)
			}
			return fmt.Errorf("error counting jobs with status %s when computing metrics. details: %w", *status, err)
		}

		err = key_value.New().Set(key, fmt.Sprintf("%d", total), 0)
		if err != nil {
			return fmt.Errorf("error setting value for key %s when computing metrics. details: %w", key, err)
		}

		return nil
	}
}

// strPtr is a helper function that returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}
