package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	logger "github.com/yaanno/worker"
)

type BackendMessage struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

type WorkerMessage struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

// Handler for consuming messages from the backend
type BackendMessageHandler struct {
	producer *nsq.Producer
	logger   *zerolog.Logger
}

func (h *BackendMessageHandler) HandleMessage(m *nsq.Message) error {
	var msg BackendMessage
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal message")
		return err
	}
	h.logger.Info().Interface("Received message from backend:", &msg).Send()
	wmsg, err := json.Marshal(&WorkerMessage{Body: "Hello from worker", ID: msg.ID})
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal message")
		return err
	}
	h.sendResponse(wmsg)
	return nil
}

func (h *BackendMessageHandler) sendResponse(wmsg []byte) {
	h.logger.Info().Msg("Sending message to backend")
	err := h.producer.Publish("worker_to_backend", []byte(wmsg))
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to publish response")
	}
	h.logger.Info().Msg("Message sent to backend")
}

func main() {
	logger := logger.InitLogger("worker", "development")
	// Signal handling for graceful stopping
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	logger.Info().Msg("Starting Backend...")

	// Create a single producer instance
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("nsqd:4150", config)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create producer")
		<-stop
		return
	}
	defer producer.Stop()

	// Create the consumer and set the handler
	consumerConfig := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("backend_to_worker", "channel1", consumerConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create consumer")
		<-stop
		return
	}

	handler := &BackendMessageHandler{producer: producer, logger: &logger}
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
