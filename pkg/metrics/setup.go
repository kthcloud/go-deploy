package metrics

import "github.com/prometheus/client_golang/prometheus"

func Setup() {
	prometheus.MustRegister(UniqueUsers)
	prometheus.MustRegister(UniqueUsersToday)
	prometheus.MustRegister(UniqueUsersThisWeek)
	prometheus.MustRegister(UniqueUsersThisMonth)
	prometheus.MustRegister(UniqueUsersThisYear)
	prometheus.MustRegister(UniqueUsersYesterday)
}
