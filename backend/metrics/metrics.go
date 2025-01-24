package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MessagesSent tracks the total number of messages sent to workers
	MessagesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "backend_messages_sent_total",
		Help: "Total number of messages sent to workers",
	}, []string{"status"})

	// MessagesReceived tracks the total number of messages received from workers
	MessagesReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "backend_messages_received_total",
		Help: "Total number of messages received from workers",
	}, []string{"status"})

	// MessageProcessingDuration tracks the time spent processing messages
	MessageProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "backend_message_processing_duration_seconds",
		Help:    "Time spent processing messages",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	// NSQConnectionStatus tracks the connection status with NSQ
	NSQConnectionStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backend_nsq_connection_status",
		Help: "Current NSQ connection status (0: Disconnected, 1: Connected)",
	})

	// MessageQueueDepth tracks the current queue depth
	MessageQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backend_message_queue_depth",
		Help: "Current message queue depth",
	})
)
