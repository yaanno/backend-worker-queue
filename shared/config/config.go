package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

// ServiceConfig represents the core configuration for a service
type ServiceConfig struct {
	ServiceName string
	Environment string
	LogLevel    zerolog.Level
	Debug       bool

	// NSQ Configuration
	NSQ NSQConfig

	// Metrics Configuration
	Metrics MetricsConfig

	// Tracing Configuration
	Tracing TracingConfig

	// Retry Configuration
	Retry RetryConfig

	Messaging MessagingConfig

	Api ApiConfig
}

// MessagingConfig defines messaging configuration
type MessagingConfig struct {
	Enabled         bool
	MessageInterval time.Duration
}

// NSQConfig contains NSQ-specific configuration
type NSQConfig struct {
	Address         string
	Channels        ChannelConfig
	ConnectionRetry RetryConfig
}

// ChannelConfig defines NSQ channel configurations
type ChannelConfig struct {
	FromWorker string
	ToWorker   string
	Name       string
}

// MetricsConfig defines metrics server configuration
type MetricsConfig struct {
	Enabled bool
	Port    int
}

// TracingConfig defines distributed tracing configuration
type TracingConfig struct {
	Enabled    bool
	Provider   string
	SampleRate float64
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

type ApiConfig struct {
	RateLimit      int
	BaseURL        string
	RetryLimit     int
	RequestTimeout time.Duration
	WorkerPoolSize int
}

// DefaultServiceConfig provides a base configuration with sensible defaults
func DefaultServiceConfig(serviceName string) *ServiceConfig {
	return &ServiceConfig{
		ServiceName: serviceName,
		Environment: getEnv("ENV", "development"),
		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", "info")),
		Debug:       getEnvBool("DEBUG", false),

		NSQ: NSQConfig{
			Address: getEnv("NSQ_ADDRESS", "localhost:4150"),
			Channels: ChannelConfig{
				FromWorker: getEnv("NSQ_FROM_WORKER_CHANNEL", "from-worker"),
				ToWorker:   getEnv("NSQ_TO_WORKER_CHANNEL", "to-worker"),
				Name:       getEnv("NSQ_CHANNEL_NAME", serviceName),
			},
			ConnectionRetry: RetryConfig{
				MaxAttempts:    getEnvInt("NSQ_MAX_RETRY", 3),
				InitialBackoff: getEnvDuration("NSQ_INITIAL_BACKOFF", 100*time.Millisecond),
				MaxBackoff:     getEnvDuration("NSQ_MAX_BACKOFF", 10*time.Second),
				BackoffFactor:  getEnvFloat("NSQ_BACKOFF_FACTOR", 2.0),
			},
		},

		Metrics: MetricsConfig{
			Enabled: getEnvBool("METRICS_ENABLED", true),
			Port:    getEnvInt("METRICS_PORT", 9090),
		},

		Tracing: TracingConfig{
			Enabled:    getEnvBool("TRACING_ENABLED", false),
			Provider:   getEnv("TRACING_PROVIDER", ""),
			SampleRate: getEnvFloat("TRACING_SAMPLE_RATE", 0.1),
		},

		Retry: RetryConfig{
			MaxAttempts:    getEnvInt("GLOBAL_MAX_RETRY", 3),
			InitialBackoff: getEnvDuration("GLOBAL_INITIAL_BACKOFF", 100*time.Millisecond),
			MaxBackoff:     getEnvDuration("GLOBAL_MAX_BACKOFF", 10*time.Second),
			BackoffFactor:  getEnvFloat("GLOBAL_BACKOFF_FACTOR", 2.0),
		},

		Messaging: MessagingConfig{
			Enabled:         getEnvBool("MESSAGING_ENABLED", false),
			MessageInterval: getEnvDuration("MESSAGING_INTERVAL", 10*time.Second),
		},

		Api: ApiConfig{
			RateLimit:      getEnvInt("API_RATE_LIMIT", 100),
			BaseURL:        getEnv("API_BASE_URL", "http://localhost:3000"),
			RetryLimit:     getEnvInt("API_RETRY_LIMIT", 3),
			RequestTimeout: getEnvDuration("API_REQUEST_TIMEOUT", 10*time.Second),
			WorkerPoolSize: getEnvInt("API_WORKER_POOL_SIZE", 10),
		},
	}
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// Validate checks the configuration for potential issues
func (c *ServiceConfig) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if c.NSQ.Address == "" {
		return fmt.Errorf("NSQ address cannot be empty")
	}

	if c.Retry.MaxAttempts < 0 {
		return fmt.Errorf("retry max attempts must be non-negative")
	}

	if c.Tracing.SampleRate < 0 || c.Tracing.SampleRate > 1 {
		return fmt.Errorf("tracing sample rate must be between 0 and 1")
	}

	return nil
}
