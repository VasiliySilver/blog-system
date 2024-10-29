package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "posts_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "posts_request_duration_seconds",
			Help:    "Duration of gRPC requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	PostsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "posts_created_total",
			Help: "Total number of created posts",
		},
	)

	DatabaseErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "posts_database_errors_total",
			Help: "Total number of database errors",
		},
	)
)
