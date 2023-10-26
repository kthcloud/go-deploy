package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	UniqueUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_unique_users",
		Help: "Total unique users",
	})
)
