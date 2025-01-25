package logger

import (
	"github.com/rs/zerolog"
	sharedLogger "github.com/yaanno/shared/logger"
)

// InitLogger creates a new logger using the shared package
func InitLogger(serviceName, environment string) zerolog.Logger {
	// Create logger config using shared package defaults
	loggerConfig := sharedLogger.DefaultLoggerConfig(serviceName)
	
	// Customize logger configuration
	loggerConfig.Environment = environment
	loggerConfig.LogLevel = zerolog.InfoLevel
	
	// Optional: Enable file logging for production
	if environment == "production" {
		loggerConfig.WriteToFile = true
		loggerConfig.FilePath = "/var/log/backend/app.log"
	}
	
	// Setup global logger configuration
	sharedLogger.SetupGlobalLogger(loggerConfig)
	
	// Create and return the configured logger
	return sharedLogger.NewLogger(loggerConfig)
}
