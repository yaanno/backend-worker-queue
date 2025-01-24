package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Retry limit
	retryLimit := 5

	// Initialize messaging
	messaging := message.NewMessaging(config, &logger)

	// Initialize the worker pool
	workerPool := pool.NewWorkerPool(1, 5)

	// Initialize the HTTP client
	httpClient := &http.Client{}

	// Initialize the worker service with the messaging instance and worker pool
	workerService := service.NewWorkerService(workerPool, messaging, &logger, httpClient, retryLimit)
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

	// Monitoring
	go func() {
		for {
			time.Sleep(5 * time.Second)
			consumers := messaging.Monitor()
			logger.Info().Msg("Consumer Status:")
			for _, consumer := range consumers {
				logger.Info().Msgf(
					"Received: %d, Processed: %d, Requeued: %d",
					consumer.Stats().MessagesReceived,
					consumer.Stats().MessagesFinished,
					consumer.Stats().MessagesRequeued,
				)
			}
			tasks := workerService.GetTaskCount()
			poolSize := workerService.GetPoolSize()
			logger.Info().Msgf("Task Count: %d, Pool Size: %d", tasks, poolSize)
			workers := workerService.Monitor()
			logger.Info().Msg("Worker Status:")
			for _, worker := range workers {
				logger.Info().Msgf(
					"Worker ID: %d, State: %s, Elapsed Time: %s",
					worker.ID,
					worker.State,
					worker.ElapsedTime,
				)
			}
		}
	}()

	// Wait for stop signal
	<-stop

	// Shut down messaging and worker service
	messaging.ShutDown()
	workerService.Stop()
}
