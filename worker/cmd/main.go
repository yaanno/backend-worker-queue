package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/yaanno/worker/config"
	"github.com/yaanno/worker/handler"
	"github.com/yaanno/worker/logger"
	"github.com/yaanno/worker/message"
	"github.com/yaanno/worker/metrics"
	"github.com/yaanno/worker/service"
)

func main() {
	// Load configuration (replace with your actual configuration loading logic)
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := logger.InitLogger("worker", "development")

	// Initialize metrics
	metrics.InitMetrics()

	// Set up a channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info().Msg("Starting Worker...")

	// Start metrics server
	go func() {
		metricsServer := &http.Server{
			Addr:    ":2112", // Prometheus default metrics port
			Handler: promhttp.Handler(),
		}
		logger.Info().Msg("Starting metrics server on :2112")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Metrics server error")
		}
	}()

	messaging, err := createMessaging(config, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create messaging client")
	}
	defer messaging.ShutDown()

	workerService, err := createWorkerService(config, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create worker service")
	}
	defer workerService.Stop()

	responseHandler := handler.NewMessageResponseHandler(ctx, &logger, workerService, messaging)

	err = messaging.Initialize(responseHandler)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize messaging")
	}

	// Monitoring
	go func() {
		for {
			time.Sleep(1 * time.Second)
			stats := messaging.GetStats()
			logger.Info().Interface("stats", stats).Send()
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			workerService.Monitor()
		}
	}()

	// Wait for stop signal
	<-stop
}

func loadConfig() (*config.Config, error) {
	cfg := config.NewConfig()
	// Add any additional configuration loading logic here if needed
	return cfg, nil
}

func createWorkerService(config *config.Config, logger *zerolog.Logger) (*service.WorkerService, error) {
	workerPool, err := ants.NewPool(config.Api.WorkerPoolSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create worker pool: %w", err)
	}

	httpClient := &http.Client{
		Timeout: config.Api.RequestTimeout,
	}

	return service.NewWorkerService(workerPool, logger, httpClient, service.APIConfig{
		RateLimit: config.Api.RateLimit,
		BaseURL:   config.Api.BaseURL,
	}, config.Api.RetryLimit), nil
}

func createMessaging(config *config.Config, logger *zerolog.Logger) (*message.Messaging, error) {
	messaging := message.NewMessaging(config, logger)
	return messaging, nil
}
