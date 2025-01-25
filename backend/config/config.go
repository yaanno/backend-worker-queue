package config

import (
	"github.com/yaanno/shared/config"
)

// Config is now a type alias to the shared configuration
type Config = config.ServiceConfig

// NewConfig creates a new configuration using the shared package defaults
func NewConfig() *Config {
	cfg := config.DefaultServiceConfig("backend")

	// Optional: Add backend-specific configuration overrides
	cfg.NSQ.Channels.FromWorker = "worker-to-backend"
	cfg.NSQ.Channels.ToWorker = "backend-to-worker"
	cfg.NSQ.Channels.Name = "common"

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	return cfg
}

// Additional backend-specific helper functions can be added here if needed
