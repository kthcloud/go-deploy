package metrics

import (
	"fmt"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"go-deploy/models/sys/key_value"
	"go-deploy/utils"
	"log"
	"strconv"
)

type MetricDefinition struct {
	Name        string
	Description string
	Key         string
	MetricType  ginmetrics.MetricType
}

const (
	KeyUsersTotal         = "metrics:users:total"
	KeyDailyActiveUsers   = "metrics:users:daily_active"
	KeyMonthlyActiveUsers = "metrics:users:monthly_active"

	KeyJobsTotal      = "metrics:jobs:total"
	KeyJobsPending    = "metrics:jobs:pending"
	KeyJobsRunning    = "metrics:jobs:running"
	KeyJobsFailed     = "metrics:jobs:failed"
	KeyJobsTerminated = "metrics:jobs:terminated"
	KeyJobsCompleted  = "metrics:jobs:completed"
)

func Setup() {
	collectors := GetCollectors()

	m := ginmetrics.GetMonitor()

	for _, def := range collectors {
		switch def.MetricType {
		case ginmetrics.Gauge:
			err := m.AddMetric(&ginmetrics.Metric{
				Type:        ginmetrics.Gauge,
				Name:        def.Name,
				Description: def.Description,
				Labels:      []string{},
			})
			if err != nil {
				log.Fatalln("failed to add metric", def.Name, "to monitor. details:", err)
			}
		}
	}
}

func Sync() {
	client := key_value.New()
	collectors := GetCollectors()
	monitor := ginmetrics.GetMonitor()

	for _, collector := range collectors {
		valueStr, err := client.Get(collector.Key)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error getting value for key %s when synchronizing metrics. details: %w", collector.Key, err))
			continue
		}

		if valueStr == "" {
			continue
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error parsing value %s when synchronizing metrics. details: %w", valueStr, err))
			continue
		}

		metric := monitor.GetMetric(collector.Name)
		if metric == nil {
			utils.PrettyPrintError(fmt.Errorf("metric %s not found when synchronizing metrics", collector.Name))
			continue
		}

		switch collector.MetricType {
		case ginmetrics.Gauge:
			err = metric.SetGaugeValue([]string{}, value)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error setting gauge value for metric %s when synchronizing metrics. details: %w", collector.Name, err))
				return
			}
		}
	}
}

func GetCollectors() []MetricDefinition {
	return []MetricDefinition{
		{
			Name:        "users_total",
			Description: "Total number of users",
			Key:         KeyUsersTotal,
			MetricType:  ginmetrics.Gauge,
		},
		{
			Name:        "users_daily_active",
			Description: "Number of users active every day the last 2 days",
			Key:         KeyDailyActiveUsers,
			MetricType:  ginmetrics.Gauge,
		},
		{
			Name:        "users_monthly_active",
			Description: "Number of users active every month the last 2 months",
			Key:         KeyMonthlyActiveUsers,
			MetricType:  ginmetrics.Gauge,
		},
		{
			Name:        "jobs_total",
			Description: "Total number of jobs",
			Key:         KeyJobsTotal,
			MetricType:  ginmetrics.Gauge,
		},
		{
			Name:        "jobs_pending",
			Description: "Number of jobs pending",
			Key:         KeyJobsPending,
			MetricType:  ginmetrics.Gauge,
		},
		{
			Name:        "jobs_running",
			Description: "Number of jobs running",
			Key:         KeyJobsRunning,
			MetricType:  ginmetrics.Gauge,
		},
		{
			Name:        "jobs_failed",
			Description: "Number of jobs failed",
			Key:         KeyJobsFailed,
			MetricType:  ginmetrics.Gauge,
		},
		{
			Name:        "jobs_terminated",
			Description: "Number of jobs terminated",
			Key:         KeyJobsTerminated,
			MetricType:  ginmetrics.Gauge,
		},
		{
			Name:        "jobs_completed",
			Description: "Number of jobs completed",
			Key:         KeyJobsCompleted,
			MetricType:  ginmetrics.Gauge,
		},
	}
}
