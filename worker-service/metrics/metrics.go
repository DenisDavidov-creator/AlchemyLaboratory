package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	MessageProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_message_processed_total",
			Help: "Total processed messages",
		},
		[]string{"status"},
	)

	RetriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_retries_total",
			Help: "Total worker retries",
		},
		[]string{},
	)
	DlqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_dlq_total",
			Help: "Total worker DLQ",
		},
		[]string{},
	)

	JobsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_jobs_in_flight",
			Help: "Jobs being processed right now",
		},
	)

	PoolSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_pool_size",
			Help: "Configured worker pool capacity",
		},
	)

	BatchSize = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "worker_batch_size",
			Help:    "Records per polled batch",
			Buckets: prometheus.ExponentialBuckets(1, 2, 12),
		},
	)

	BatchDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "worker_batch_duration_seconds",
			Help:    "Time spent processing one batch, from poll to last job",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 12),
		},
	)
)

func init() {
	prometheus.MustRegister(
		MessageProcessedTotal,
		RetriesTotal,
		DlqTotal,
		JobsInFlight,
		PoolSize,
		BatchSize,
		BatchDuration,
	)
}
