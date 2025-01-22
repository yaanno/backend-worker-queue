package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/yaanno/backend/handler"
	logger "github.com/yaanno/backend/logger"
	"github.com/yaanno/backend/message"
	"github.com/yaanno/backend/model"
)

type WorkerMessage struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

func main() {
	// Signal handling for graceful stopping
	logger := logger.InitLogger("backend", "development")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	logger.Info().Msg("Starting Backend...")

	config := nsq.NewConfig()

	handler := handler.NewWorkerResponseHandler(&logger)

	messaging := message.NewMessaging(config, &logger)

	messaging.Initialize(handler)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				id := uuid.New()
				msg := model.BackendMessage{Body: "Hello from backend", ID: int(id.ID())}
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
