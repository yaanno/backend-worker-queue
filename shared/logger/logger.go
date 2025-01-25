package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// LoggerConfig defines configuration for logger initialization
type LoggerConfig struct {
	ServiceName string
	Environment string
	LogLevel    zerolog.Level
	EnableColor bool
	WriteToFile bool
	FilePath    string
}

// NewLogger creates a configured zerolog logger
func NewLogger(config LoggerConfig) zerolog.Logger {
	// Set global log level
	zerolog.SetGlobalLevel(config.LogLevel)

	// Configure error stack marshaling
	// zerolog.ErrorStackMarshaler =

	// Create base logger
	var logger zerolog.Logger

	// Configure output
	if config.EnableColor {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		logger = zerolog.New(os.Stdout)
	}

	// Add file logging if enabled
	var logOutput zerolog.Logger
	if config.WriteToFile && config.FilePath != "" {
		logFile, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err == nil {
			logOutput = logger.Output(logFile)
		} else {
			logger.Warn().Err(err).Msg("Failed to open log file, falling back to stdout")
			logOutput = logger
		}
	} else {
		logOutput = logger
	}

	// Add context to logger
	return logOutput.With().
		Str("service", config.ServiceName).
		Str("environment", config.Environment).
		Timestamp().
		Logger()
}

// WithContext adds additional context to an existing logger
func WithContext(logger zerolog.Logger, ctx map[string]interface{}) zerolog.Logger {
	logCtx := logger.With()
	for k, v := range ctx {
		switch val := v.(type) {
		case string:
			logCtx = logCtx.Str(k, val)
		case int:
			logCtx = logCtx.Int(k, val)
		case int64:
			logCtx = logCtx.Int64(k, val)
		case float64:
			logCtx = logCtx.Float64(k, val)
		case bool:
			logCtx = logCtx.Bool(k, val)
		default:
			logCtx = logCtx.Interface(k, v)
		}
	}
	return logCtx.Logger()
}

// DefaultLoggerConfig provides a base configuration
func DefaultLoggerConfig(serviceName string) LoggerConfig {
	return LoggerConfig{
		ServiceName: serviceName,
		Environment: "development",
		LogLevel:    zerolog.InfoLevel,
		EnableColor: true,
		WriteToFile: false,
		FilePath:    "",
	}
}

// SetupGlobalLogger configures the global zerolog configuration
func SetupGlobalLogger(config LoggerConfig) {
	zerolog.SetGlobalLevel(config.LogLevel)
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
}
