package config

import (
	"time"

	"github.com/yaanno/shared/config"
)

// Config is now a type alias to the shared configuration
type Config = config.ServiceConfig

// NewConfig creates a new configuration using the shared package defaults
func NewConfig() *Config {
	cfg := config.DefaultServiceConfig("worker")

	// Optional: Add worker-specific configuration overrides
	cfg.NSQ.Channels.FromWorker = "worker-to-backend"
	cfg.NSQ.Channels.ToWorker = "backend-to-worker"
	cfg.NSQ.Channels.Name = "common"

	// Configure worker-specific parameters
	cfg.Retry.MaxAttempts = 5 // More retries for worker
	cfg.Retry.InitialBackoff = 200 * time.Millisecond

	// Add worker-specific service configurations
	cfg.Metrics.Port = 2112 // Different metrics port

	cfg.Api.BaseURL = "https://potterapi-fedeperin.vercel.app/es/characters?search=Weasley"

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	return cfg
}

// Additional worker-specific helper functions can be added here if needed
