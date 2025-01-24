package retry

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomFactor    float64
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:      3,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		RandomFactor:    0.1,
	}
}

var (
	retryAttempts = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "retry_attempts_total",
			Help:    "Number of retry attempts made",
			Buckets: []float64{0, 1, 2, 3, 4, 5},
		},
		[]string{"operation"},
	)

	retryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "retry_duration_seconds",
			Help:    "Time spent retrying operations",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 5),
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(retryAttempts, retryDuration)
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// DoWithRetry executes the given function with exponential backoff retry logic
func DoWithRetry(ctx context.Context, operation string, logger *zerolog.Logger, config *RetryConfig, fn RetryableFunc) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var err error
	currentInterval := config.InitialInterval
	timer := prometheus.NewTimer(retryDuration.WithLabelValues(operation))
	defer timer.ObserveDuration()

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Record the attempt
		retryAttempts.WithLabelValues(operation).Observe(float64(attempt))

		// Try the operation
		err = fn()
		if err == nil {
			return nil
		}

		// If this was our last attempt, return the error
		if attempt == config.MaxRetries {
			return fmt.Errorf("failed after %d attempts: %w", attempt+1, err)
		}

		// Calculate next interval with jitter
		jitter := 1.0 + (rand.Float64()*2-1.0)*config.RandomFactor
		nextInterval := time.Duration(float64(currentInterval) * config.Multiplier * jitter)

		// Ensure we don't exceed MaxInterval
		if nextInterval > config.MaxInterval {
			nextInterval = config.MaxInterval
		}

		logger.Debug().
			Str("operation", operation).
			Int("attempt", attempt+1).
			Dur("next_interval", nextInterval).
			Err(err).
			Msg("Retrying operation after error")

		// Wait for either the backoff duration or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(nextInterval):
			currentInterval = nextInterval
		}
	}

	return err
}
