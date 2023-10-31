package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	UniqueUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_unique_users",
		Help: "Total unique users",
	})
	UniqueUsersToday = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "unique_users_today",
		Help: "Unique users today",
	})
	UniqueUsersThisWeek = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "unique_users_this_week",
		Help: "Unique users this week",
	})
	UniqueUsersThisMonth = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "unique_users_this_month",
		Help: "Unique users this month",
	})
	UniqueUsersThisYear = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "unique_users_this_year",
		Help: "Unique users this year",
	})
	UniqueUsersYesterday = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "unique_users_yesterday",
		Help: "Unique users yesterday",
	})
)
