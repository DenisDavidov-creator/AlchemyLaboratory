package server

import "github.com/prometheus/client_golang/prometheus"

var (
	httpRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"methods", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP reqest duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"methods", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestTotal, httpRequestDuration)
}
