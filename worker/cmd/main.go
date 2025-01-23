package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
	handler "github.com/yaanno/worker/handler"
	logger "github.com/yaanno/worker/logger"
	"github.com/yaanno/worker/message"
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

	// Initialize the backend response handler with the messaging instance
	handler := handler.NewMessageResponseHandler(&logger)

	err := messaging.Initialize(handler)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize messaging")
		return
	}

	// Start listening for messages in a goroutine
	go messaging.ListenForMessages(ctx)

	// Wait for stop signal
	<-stop

	// Shut down messaging gracefully
	messaging.ShutDown()
}
