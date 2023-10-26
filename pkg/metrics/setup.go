package metrics

import "github.com/prometheus/client_golang/prometheus"

func Setup() {
	prometheus.MustRegister(UniqueUsers)
}
