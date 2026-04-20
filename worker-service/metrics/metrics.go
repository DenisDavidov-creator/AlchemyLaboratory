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
)

func init() {
	prometheus.MustRegister(MessageProcessedTotal, RetriesTotal, DlqTotal)
}
