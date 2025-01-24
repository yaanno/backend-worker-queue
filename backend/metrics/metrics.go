package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registrationOnce sync.Once

	// MessageProcessingDuration tracks the time it takes to process messages
	MessageProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "message_processing_duration_seconds",
			Help:    "Time taken to process messages",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// MessagesSent tracks the number of messages sent
	MessagesSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "messages_sent_total",
			Help: "Number of messages sent",
		},
		[]string{"status"},
	)

	// MessagesReceived tracks the number of messages received
	MessagesReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "messages_received_total",
			Help: "Number of messages received",
		},
		[]string{"status"},
	)

	// NSQConnectionStatus indicates if we're connected to NSQ
	NSQConnectionStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nsq_connection_status",
		Help: "Connection status to NSQ (1 for connected, 0 for disconnected)",
	})

	// MessageQueueDepth tracks the current depth of the message queue
	MessageQueueDepth = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "message_queue_depth",
		Help: "Current message queue depth",
	})

	// MessageSize tracks the size of messages in bytes
	MessageSize = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "message_size_bytes",
		Help:    "Size of messages in bytes",
		Buckets: []float64{64, 128, 256, 512, 1024, 2048, 4096, 8192},
	})
)

// InitMetrics registers all metrics with Prometheus.
// It's safe to call this function multiple times as it will only register metrics once.
func InitMetrics() {
	registrationOnce.Do(func() {
		// Register all metrics
		prometheus.MustRegister(
			MessageProcessingDuration,
			MessagesSent,
			MessagesReceived,
			MessageQueueDepth,
			NSQConnectionStatus,
			MessageSize,
		)
	})
}
