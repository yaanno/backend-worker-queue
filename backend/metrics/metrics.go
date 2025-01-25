package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	sharedMetrics "github.com/yaanno/shared/metrics"
)

var (
	// Common metrics from shared package
	CommonMetrics *sharedMetrics.CommonMetrics

	// Specific backend metrics
	NSQConnectionStatus       *prometheus.GaugeVec
	MessagesSent              *prometheus.CounterVec
	MessagesReceived          *prometheus.CounterVec
	MessageSize               *prometheus.HistogramVec
	MessageQueueDepth         *prometheus.GaugeVec
	MessageProcessingDuration *prometheus.HistogramVec
	WorkerPoolStatus          *prometheus.GaugeVec
	TaskProcessingTime        *prometheus.HistogramVec
	TaskCompletionRate        *prometheus.CounterVec
	WorkerUtilization         *prometheus.GaugeVec
	MessageProcessingTotal    *prometheus.CounterVec
	CircuitBreakerState       *prometheus.GaugeVec
	APIRequestDuration        *prometheus.HistogramVec
)

// InitMetrics initializes all metrics using the shared metrics registry
func InitMetrics() {
	// Create metrics registry with default configuration
	metricsConfig := sharedMetrics.DefaultMetricsConfig("backend")
	metricsRegistry := sharedMetrics.NewMetricsRegistry(metricsConfig)

	// Initialize common metrics
	CommonMetrics = metricsRegistry.CreateCommonMetrics()

	// Create backend-specific metrics
	NSQConnectionStatus = metricsRegistry.CreateGauge(
		"nsq_connection_status",
		"NSQ connection status (0 = disconnected, 1 = connected)",
		"type",
	)

	MessagesSent = metricsRegistry.CreateCounter(
		"messages_sent_total",
		"Total number of messages sent",
		"status",
	)

	MessagesReceived = metricsRegistry.CreateCounter(
		"messages_received_total",
		"Total number of messages received",
		"status",
	)

	MessageSize = metricsRegistry.CreateHistogram(
		"message_size_bytes",
		"Size of messages in bytes",
		prometheus.ExponentialBuckets(64, 2, 10),
	)

	MessageQueueDepth = metricsRegistry.CreateGauge(
		"message_queue_depth",
		"Current depth of the message queue",
		"queue",
	)

	MessageProcessingDuration = metricsRegistry.CreateHistogram(
		"message_processing_duration_seconds",
		"Duration of message processing",
		prometheus.DefBuckets,
		"operation",
	)

	CircuitBreakerState = metricsRegistry.CreateGauge(
		"circuit_breaker_state",
		"Circuit breaker state (0 = closed, 1 = open)",
		"operation",
	)

	APIRequestDuration = metricsRegistry.CreateHistogram(
		"api_request_duration_seconds",
		"Duration of API requests",
		prometheus.DefBuckets,
		"operation",
	)

	// Start metrics server
	if err := metricsRegistry.StartMetricsServer(); err != nil {
		panic(err)
	}
}
