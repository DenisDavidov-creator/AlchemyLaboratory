package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	GrpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_request_total",
			Help: "Total number of gRPC reuqests",
		},
		[]string{"method", "status"},
	)

	GrpcRequestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration",
			Help:    "gRPC duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(GrpcRequestsDuration, GrpcRequestsTotal)
}
