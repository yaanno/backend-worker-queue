package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
)

// Handler for consuming messages from the worker
type WorkerResponseHandler struct{}

func (h *WorkerResponseHandler) HandleMessage(m *nsq.Message) error {
	log.Printf("Received message from worker: %s", string(m.Body))
	return nil
}

func main() {
	// Signal handling for graceful stopping
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	time.Sleep(10 * time.Second)
	log.Println("Starting Backend...")

	// Producer code to send messages to the worker in a loop
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("nsqd:4150", config)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Stop()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err = producer.Publish("backend_to_worker", []byte("Message from Backend"))
				if err != nil {
					log.Printf("Failed to publish message: %v", err)
				} else {
					log.Println("Message sent to worker")
				}
			case <-stop:
				log.Println("Stopping message loop")
				return
			}
		}
	}()

	// Consumer code to receive messages from the worker
	consumer, err := nsq.NewConsumer("worker_to_backend", "channel1", config)
	if err != nil {
		log.Fatal(err)
	}

	consumer.AddHandler(&WorkerResponseHandler{})

	err = consumer.ConnectToNSQD("nsqd:4150")
	if err != nil {
		log.Fatal(err)
	}

	// Wait for stop signal
	<-stop
	log.Println("Shutting down gracefully...")
	consumer.Stop()
}
