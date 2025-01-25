package metrics

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsConfig defines configuration for metrics collection
type MetricsConfig struct {
	ServiceName     string
	Namespace       string
	SubsystemPrefix string
	Port            int
}

// MetricsRegistry manages Prometheus metrics
type MetricsRegistry struct {
	config         MetricsConfig
	registeredOnce sync.Once
	registry       *prometheus.Registry
}

// NewMetricsRegistry creates a new metrics registry
func NewMetricsRegistry(config MetricsConfig) *MetricsRegistry {
	return &MetricsRegistry{
		config:   config,
		registry: prometheus.NewRegistry(),
	}
}

// StartMetricsServer starts a Prometheus metrics server
func (m *MetricsRegistry) StartMetricsServer() error {
	http.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		Registry: m.registry,
	}))

	addr := fmt.Sprintf(":%d", m.config.Port)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Printf("Failed to start metrics server: %v\n", err)
		}
	}()

	return nil
}

// CreateCounter creates a Prometheus counter metric
func (m *MetricsRegistry) CreateCounter(name, help string, labels ...string) *prometheus.CounterVec {
	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.getSubsystem(),
			Name:      name,
			Help:      help,
		},
		labels,
	)
	m.registry.MustRegister(counter)
	return counter
}

// CreateHistogram creates a Prometheus histogram metric
func (m *MetricsRegistry) CreateHistogram(name, help string, buckets []float64, labels ...string) *prometheus.HistogramVec {
	histogram := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.getSubsystem(),
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)
	m.registry.MustRegister(histogram)
	return histogram
}

// CreateGauge creates a Prometheus gauge metric
func (m *MetricsRegistry) CreateGauge(name, help string, labels ...string) *prometheus.GaugeVec {
	gauge := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.getSubsystem(),
			Name:      name,
			Help:      help,
		},
		labels,
	)
	m.registry.MustRegister(gauge)
	return gauge
}

// getSubsystem returns the subsystem name with optional prefix
func (m *MetricsRegistry) getSubsystem() string {
	if m.config.SubsystemPrefix != "" {
		return fmt.Sprintf("%s_%s", m.config.SubsystemPrefix, m.config.ServiceName)
	}
	return m.config.ServiceName
}

// DefaultMetricsConfig provides a base configuration
func DefaultMetricsConfig(serviceName string) MetricsConfig {
	return MetricsConfig{
		ServiceName:     serviceName,
		Namespace:       "yaanno",
		SubsystemPrefix: "service",
		Port:            9090,
	}
}

// CommonMetrics provides a set of standard metrics for services
type CommonMetrics struct {
	RequestCounter     *prometheus.CounterVec
	RequestDuration    *prometheus.HistogramVec
	ActiveConnections  *prometheus.GaugeVec
	ErrorCounter       *prometheus.CounterVec
}

// CreateCommonMetrics initializes a set of common service metrics
func (m *MetricsRegistry) CreateCommonMetrics() *CommonMetrics {
	return &CommonMetrics{
		RequestCounter: m.CreateCounter(
			"requests_total", 
			"Total number of requests processed", 
			"method", "status",
		),
		RequestDuration: m.CreateHistogram(
			"request_duration_seconds", 
			"Request processing time in seconds", 
			prometheus.DefBuckets,
			"method", "status",
		),
		ActiveConnections: m.CreateGauge(
			"active_connections", 
			"Number of active connections", 
			"type",
		),
		ErrorCounter: m.CreateCounter(
			"errors_total", 
			"Total number of errors", 
			"type", "code",
		),
	}
}
