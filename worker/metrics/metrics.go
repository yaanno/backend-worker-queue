package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	sharedMetrics "github.com/yaanno/shared/metrics"
)

var (
	// Common metrics from shared package
	CommonMetrics *sharedMetrics.CommonMetrics

	// Specific worker metrics
	WorkerPoolStatus          *prometheus.GaugeVec
	TaskProcessingTime        *prometheus.HistogramVec
	TaskCompletionRate        *prometheus.CounterVec
	WorkerUtilization         *prometheus.GaugeVec
	MessageProcessingDuration *prometheus.HistogramVec
	MessageProcessingTotal    *prometheus.CounterVec
	CircuitBreakerState       prometheus.Gauge // Track circuit breaker state: 0=Open, 1=HalfOpen, 2=Closed
	APIRequestDuration        *prometheus.HistogramVec
)

// InitMetrics initializes all metrics using the shared metrics registry
func InitMetrics() {
	// Create metrics registry with default configuration
	metricsConfig := sharedMetrics.DefaultMetricsConfig("worker")
	metricsRegistry := sharedMetrics.NewMetricsRegistry(metricsConfig)

	// Initialize common metrics
	CommonMetrics = metricsRegistry.CreateCommonMetrics()

	// Create worker-specific metrics
	WorkerPoolStatus = metricsRegistry.CreateGauge(
		"worker_pool_status",
		"Status of worker pool (active workers, max workers)",
		"type",
	)

	TaskProcessingTime = metricsRegistry.CreateHistogram(
		"task_processing_duration_seconds",
		"Duration of task processing",
		prometheus.DefBuckets,
		"task_type",
	)

	TaskCompletionRate = metricsRegistry.CreateCounter(
		"task_completion_total",
		"Total number of tasks completed",
		"status", "type",
	)

	WorkerUtilization = metricsRegistry.CreateGauge(
		"worker_utilization",
		"Current worker pool utilization",
		"pool_name",
	)

	MessageProcessingDuration = metricsRegistry.CreateHistogram(
		"message_processing_duration_seconds",
		"Duration of message processing",
		prometheus.DefBuckets,
		"type",
	)

	MessageProcessingTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_message_processing_total",
			Help: "Total number of messages processed by status",
		},
		[]string{"status"},
	)
	prometheus.MustRegister(MessageProcessingTotal)

	CircuitBreakerState = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "circuit_breaker_state",
		Help: "State of the circuit breaker: 0=Open, 1=HalfOpen, 2=Closed",
	})
	prometheus.MustRegister(CircuitBreakerState)

	APIRequestDuration = metricsRegistry.CreateHistogram(
		"api_request_duration_seconds",
		"Duration of API requests",
		prometheus.DefBuckets,
		"type",
	)

	// Start metrics server
	if err := metricsRegistry.StartMetricsServer(); err != nil {
		panic(err)
	}
}
