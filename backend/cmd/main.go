package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yaanno/backend/config"
	"github.com/yaanno/backend/handler"
	logger "github.com/yaanno/backend/logger"
	"github.com/yaanno/backend/message"
	"github.com/yaanno/backend/model"
)

func main() {
	// Initialize logger and config
	logger := logger.InitLogger("backend", "development")
	config := config.NewConfig()

	// Signal handling for graceful stopping
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	logger.Info().Msg("Starting Backend...")

	// Create handler and messaging
	handler := handler.NewMessageResponseHandler(&logger)
	messaging := message.NewMessaging(config, &logger)

	if err := messaging.Initialize(handler); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize messaging")
	}

	// Start metrics server
	go func() {
		metricsServer := &http.Server{
			Addr:    fmt.Sprintf(":%s", config.MetricsPort),
			Handler: promhttp.Handler(),
		}
		logger.Info().Msgf("Starting metrics server on :%s", config.MetricsPort)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Metrics server error")
		}
	}()

	// Start message publishing loop
	go func() {
		ticker := time.NewTicker(config.MessageInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				id := uuid.New()
				msg := model.Response{Body: "Hello from backend", ID: int(id.ID())}
				err := messaging.PublishMessage(&msg)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to publish message")
				} else {
					logger.Info().Msg("Message sent to worker")
				}
			case <-stop:
				logger.Info().Msg("Stopping message loop")
				return
			}
		}
	}()

	// Wait for stop signal
	<-stop
	logger.Info().Msg("Shutting down gracefully...")
	messaging.ShutDown()
}
