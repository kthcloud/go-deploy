package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	Metrics MetricsType
)

type MetricsType struct {
	Registry *prometheus.Registry
	Gauges   map[string]prometheus.Gauge
}
