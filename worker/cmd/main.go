package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/panjf2000/ants/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	handler "github.com/yaanno/worker/handler"
	logger "github.com/yaanno/worker/logger"
	"github.com/yaanno/worker/message"
	"github.com/yaanno/worker/service"
)

type Config struct {
	NSQConfig        *nsq.Config
	WorkerPoolSize   int
	GlobalRetryLimit int
	APIConfig        APIConfig
	Timeouts         TimeoutConfig
}

type APIConfig struct {
	BaseURL       string
	APIKey        string
	RateLimit     int // Requests per second
	MaxRetries    int
	APIRetryLimit int
}

type TimeoutConfig struct {
	RequestTimeout time.Duration
	TaskTimeout    time.Duration
}

func main() {
	// Load configuration (replace with your actual configuration loading logic)
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := logger.InitLogger("worker", "development")

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

	// Start listening for messages in a goroutine
	// go func() {
	// 	defer cancel() // Propagate cancellation to listening goroutine
	// 	err := messaging.ListenForMessages(ctx)
	// 	if err != nil {
	// 		logger.Error().Err(err).Msg("Error listening for messages")
	// 	}
	// }()

	// Monitoring
	go func() {
		for {
			time.Sleep(1 * time.Second)
			stats := messaging.GetStats()
			logger.Info().Interface("stats", stats).Send()
			// for _, consumer := range stats {
			// 	logger.Info().Interface("stats", consumer.Stats()).Send()
			// }
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

func loadConfig() (*Config, error) {
	// Implement your configuration loading logic here (e.g., read from file, environment variables)
	return &Config{
		NSQConfig:        nsq.NewConfig(), // Replace with actual configuration
		WorkerPoolSize:   10,              // Adjust as needed
		GlobalRetryLimit: 5,
		APIConfig: APIConfig{
			BaseURL: "https://potterapi-fedeperin.vercel.app/es/characters?search=Weasley",
			// APIKey:        "your-api-key",
			RateLimit:     10,
			MaxRetries:    3,
			APIRetryLimit: 5,
		},
		Timeouts: TimeoutConfig{
			RequestTimeout: 10 * time.Second,
			TaskTimeout:    30 * time.Second,
		},
	}, nil
}

func createWorkerService(config *Config, logger *zerolog.Logger) (*service.WorkerService, error) {
	workerPool, err := ants.NewPool(config.WorkerPoolSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create worker pool: %w", err)
	}

	httpClient := &http.Client{
		Timeout: config.Timeouts.RequestTimeout,
	}

	return service.NewWorkerService(workerPool, logger, httpClient, service.APIConfig{
		RateLimit: config.APIConfig.RateLimit,
		BaseURL:   config.APIConfig.BaseURL,
	}, config.GlobalRetryLimit), nil
}

func createMessaging(config *Config, logger *zerolog.Logger) (*message.Messaging, error) {
	messaging := message.NewMessaging(config.NSQConfig, logger)
	return messaging, nil
}
