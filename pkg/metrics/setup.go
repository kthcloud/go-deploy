package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go-deploy/models/sys/key_value"
	"go-deploy/utils"
	"strconv"
)

type MetricDefinition struct {
	Key        string
	MetricType string
	Collector  interface{}
}

const (
	MetricTypeGauge = "gauge"

	KeyUsersTotal         = "metrics:users:total"
	KeyDailyActiveUsers   = "metrics:users:daily_active"
	KeyMonthlyActiveUsers = "metrics:users:monthly_active"
)

func Setup() {

	collectors := GetCollectors()

	Gauges = make(map[string]prometheus.Gauge)

	for _, def := range collectors {
		switch def.MetricType {
		case MetricTypeGauge:
			prometheus.MustRegister(def.Collector.(prometheus.Gauge))
			Gauges[def.Key] = def.Collector.(prometheus.Gauge)
		}
	}
}

func Sync() {
	client := key_value.New()

	for key, gauge := range Gauges {
		valueStr, err := client.Get(key)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error getting value for key %s when synchronizing metrics. details: %w", key, err))
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

		gauge.Set(value)
	}
}

func GetCollectors() []MetricDefinition {
	return []MetricDefinition{
		{
			Key:        KeyUsersTotal,
			MetricType: MetricTypeGauge,
			Collector: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "users_total",
				Help: "Total number of users",
			}),
		},
		{
			Key:        KeyDailyActiveUsers,
			MetricType: MetricTypeGauge,
			Collector: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "users_daily_active",
				Help: "Number of users active every day the last 2 days",
			}),
		},
		{
			Key:        KeyMonthlyActiveUsers,
			MetricType: MetricTypeGauge,
			Collector: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "users_monthly_active",
				Help: "Number of users active every month the last 2 months",
			}),
		},
	}
}
