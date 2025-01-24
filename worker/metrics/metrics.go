package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MessageProcessingDuration tracks the time spent processing messages
	MessageProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "worker_message_processing_duration_seconds",
		Help:    "Time spent processing messages",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})

	// MessageProcessingTotal tracks the total number of processed messages
	MessageProcessingTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_message_processing_total",
		Help: "Total number of processed messages",
	}, []string{"status"})

	// WorkerPoolUtilization tracks the current worker pool utilization
	WorkerPoolUtilization = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "worker_pool_utilization",
		Help: "Current worker pool utilization",
	})

	// QueueDepth tracks the current queue depth
	QueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "worker_queue_depth",
		Help: "Current queue depth",
	})

	// APIRequestDuration tracks the duration of API requests
	APIRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "worker_api_request_duration_seconds",
		Help:    "Time spent making API requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})

	// CircuitBreakerState tracks the state of the circuit breaker
	CircuitBreakerState = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "worker_circuit_breaker_state",
		Help: "Current state of the circuit breaker (0: Open, 1: Half-Open, 2: Closed)",
	})
)
