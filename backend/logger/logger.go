package logger

import (
	"os"
	"runtime/debug"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger(service string, environment string) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().
		Str("service", service).
		Str("version", debug.BuildInfo{}.Main.Version).
		Logger()

	// Set global log level based on environment
	zerolog.SetGlobalLevel(zerolog.InfoLevel) // Default to Info
	if environment == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	return logger
}
