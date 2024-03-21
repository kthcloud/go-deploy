package metrics

import (
	"context"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/key_value"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/user_repo"
	"go-deploy/pkg/metrics"
	"go-deploy/pkg/workers"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// metricsWorker is a worker that updates metrics.
func metricsWorker(ctx context.Context) {
	defer workers.OnStop("metrics worker")

	metricsFuncMap := map[string]func() error{
		"users-total":          usersTotal,
		"monthly-active-users": monthlyActiveUsers,
		"daily-active-users":   dailyActiveUsers,
		"jobs-total":           jobMetrics(metrics.KeyJobsTotal, nil),
		"jobs-pending":         jobMetrics(metrics.KeyJobsPending, strPtr(model.JobStatusPending)),
		"jobs-running":         jobMetrics(metrics.KeyJobsRunning, strPtr(model.JobStatusRunning)),
		"jobs-failed":          jobMetrics(metrics.KeyJobsFailed, strPtr(model.JobStatusFailed)),
		"jobs-terminated":      jobMetrics(metrics.KeyJobsTerminated, strPtr(model.JobStatusTerminated)),
		"jobs-completed":       jobMetrics(metrics.KeyJobsCompleted, strPtr(model.JobStatusCompleted)),
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
	total, err := user_repo.New().Count()
	if err != nil {
		return fmt.Errorf("error counting distinct users when computing metrics. details: %w", err)
	}

	err = key_value.New().Set(metrics.KeyUsersTotal, fmt.Sprintf("%d", total), 0)
	if err != nil {
		return fmt.Errorf("error setting value for key %s when computing metrics. details: %w", metrics.KeyUsersTotal, err)
	}

	return nil
}

// monthlyActiveUsers computes the number of active users and stores it in the key-value store.
func monthlyActiveUsers() error {
	total, err := user_repo.New().LastAuthenticatedAfter(time.Now().AddDate(0, -1, 0)).Count()
	if err != nil {
		return fmt.Errorf("error counting monthly active users when computing metrics. details: %w", err)
	}

	err = key_value.New().Set(metrics.KeyMonthlyActiveUsers, fmt.Sprintf("%d", total), 0)
	if err != nil {
		return fmt.Errorf("error setting value for key %s when computing metrics. details: %w", metrics.KeyMonthlyActiveUsers, err)
	}

	return nil
}

// dailyActiveUsers computes the number of active users and stores it in the key-value store.
func dailyActiveUsers() error {
	total, err := user_repo.New().LastAuthenticatedAfter(time.Now().AddDate(0, 0, -1)).Count()
	if err != nil {
		return fmt.Errorf("error counting daily active users when computing metrics. details: %w", err)
	}

	err = key_value.New().Set(metrics.KeyDailyActiveUsers, fmt.Sprintf("%d", total), 0)
	if err != nil {
		return fmt.Errorf("error setting value for key %s when computing metrics. details: %w", metrics.KeyDailyActiveUsers, err)
	}

	return nil
}

// jobMetrics computes the number of jobs and stores it in the key-value store.
func jobMetrics(key string, status *string) func() error {
	return func() error {
		filter := bson.D{}
		if status != nil {
			filter = append(filter, bson.E{Key: "status", Value: *status})
		}

		total, err := job_repo.New().AddFilter(filter).Count()
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
