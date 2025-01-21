package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	logger "github.com/yaanno/backend"
)

// Handler for consuming messages from the worker
type WorkerResponseHandler struct {
	logger   *zerolog.Logger
	consumer *nsq.Consumer
}

type Message struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

type WorkerMessage struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

func (h *WorkerResponseHandler) HandleMessage(m *nsq.Message) error {
	var msg WorkerMessage
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal message")
		return err
	}
	h.logger.Info().Interface("Received message from worker:", &msg).Send()
	return nil
}

func main() {
	// Signal handling for graceful stopping
	logger := logger.InitLogger("backend", "development")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	logger.Info().Msg("Starting Backend...")

	// Producer code to send messages to the worker in a loop
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("nsqd:4150", config)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create producer")
		<-stop
		return
	}
	defer producer.Stop()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				id := uuid.New()
				msg, err := json.Marshal(&Message{Body: "Hello from backend", ID: int(id.ID())})
				logger.Info().Msg("Sending message to worker")
				if err != nil {
					logger.Error().Err(err).Msg("Failed to marshal message")
					continue
				}
				err = producer.Publish("backend_to_worker", msg)
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

	// Consumer code to receive messages from the worker
	consumer, err := nsq.NewConsumer("worker_to_backend", "channel1", config)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create consumer")
		// log.Fatal(err)
		<-stop
		return
	}

	// consumer.AddHandler(&WorkerResponseHandler{})
	handler := &WorkerResponseHandler{consumer: consumer, logger: &logger}
	consumer.AddHandler(handler)

	err = consumer.ConnectToNSQD("nsqd:4150")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to nsqd")
		<-stop
		return
	}

	// Wait for stop signal
	<-stop
	logger.Info().Msg("Shutting down gracefully...")
	consumer.Stop()
}
