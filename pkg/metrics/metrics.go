package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	Gauges map[string]prometheus.Gauge
)
