package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"net/http"

	"github.com/nsqio/go-nsq"
	handler "github.com/yaanno/worker/handler"
	logger "github.com/yaanno/worker/logger"
	"github.com/yaanno/worker/message"
	"github.com/yaanno/worker/pool"
	"github.com/yaanno/worker/service"
)

func main() {
	// Initialize logger
	logger := logger.InitLogger("worker", "development")

	// Set up a channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info().Msg("Starting Worker...")

	// Initialize NSQ config
	config := nsq.NewConfig()

	// Initialize messaging
	messaging := message.NewMessaging(config, &logger)

	// Initialize the worker pool
	workerPool := pool.NewWorkerPool(1)

	// Initialize the HTTP client
	httpClient := &http.Client{}

	// Initialize the worker service with the messaging instance and worker pool
	workerService := service.NewWorkerService(workerPool, messaging, &logger, httpClient)
	workerService.Start() // Start the worker pool

	// Initialize the backend response handler with the worker service instance
	responseHandler := handler.NewMessageResponseHandler(&logger, workerService)

	err := messaging.Initialize(responseHandler)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize messaging")
		return
	}

	// Start listening for messages in a goroutine
	go messaging.ListenForMessages(ctx)

	// Wait for stop signal
	<-stop

	// Shut down messaging and worker service
	messaging.ShutDown()
	workerService.Stop()
}
