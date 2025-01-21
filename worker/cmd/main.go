package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
)

// Handler for consuming messages from the backend
type BackendMessageHandler struct {
	producer *nsq.Producer
}

func (h *BackendMessageHandler) HandleMessage(m *nsq.Message) error {
	log.Printf("Received message from backend: %s", string(m.Body))
	// Process the message and send a response back to the backend
	h.sendResponse("Processed message by worker")
	return nil
}

func (h *BackendMessageHandler) sendResponse(response string) {
	err := h.producer.Publish("worker_to_backend", []byte(response))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Signal handling for graceful stopping
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Create a single producer instance
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Stop()

	// Create the consumer and set the handler
	consumerConfig := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("backend_to_worker", "channel1", consumerConfig)
	if err != nil {
		log.Fatal(err)
	}

	handler := &BackendMessageHandler{producer: producer}
	consumer.AddHandler(handler)

	err = consumer.ConnectToNSQD("127.0.0.1:4150")
	if err != nil {
		log.Fatal(err)
	}

	// Wait for stop signal
	<-stop
	log.Println("Shutting down gracefully...")
	consumer.Stop()
}
