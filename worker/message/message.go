package message

import (
	"encoding/json"
	"sync"

	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	"github.com/yaanno/worker/config"
	"github.com/yaanno/worker/model"
)

// const (
// 	backend_channel = "backend_to_worker"
// 	worker_channel  = "worker_to_backend"
// 	nsqd_address    = "nsqd:4150"
// 	channel         = "channel1"
// )

type Messaging struct {
	producer *nsq.Producer
	consumer *nsq.Consumer
	config   *config.Config
	logger   *zerolog.Logger
	wg       sync.WaitGroup
}

func NewMessaging(config *config.Config, logger *zerolog.Logger) *Messaging {
	return &Messaging{
		config: config,
		logger: logger,
	}
}

func (m *Messaging) Initialize(handler nsq.Handler) error {
	var err error
	if m.logger == nil {
		m.logger = &zerolog.Logger{}
	}
	m.producer, err = nsq.NewProducer(m.config.NSQ.Address, nsq.NewConfig())
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to connect to nsqd")
	}

	// Create and start the consumer in a separate function
	err = m.startConsumer(handler, m.config.NSQ.Address)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to start consumer")
		return err
	}

	return nil
}

func (m *Messaging) startConsumer(handler nsq.Handler, nsqdAddress string) error {
	m.logger.Warn().Msg("Starting new consumer...")
	consumer, err := nsq.NewConsumer(m.config.NSQ.Channels.ToWorker, m.config.NSQ.Channels.Name, nsq.NewConfig())
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to create consumer")
		return err
	}
	m.consumer = consumer
	m.consumer.AddHandler(handler)

	err = m.consumer.ConnectToNSQD(nsqdAddress)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to connect to nsqd")
		return err
	}

	return nil
}

func (m *Messaging) PublishMessage(msg *model.Response) error {
	message, err := json.Marshal(msg)
	m.logger.Info().Msg("Sending message to backend")
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to marshal message")
		return err
	}
	m.logger.Info().Msg("Publishing message to backend")
	err = m.producer.Publish(m.config.NSQ.Channels.FromWorker, message)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to publish message")
		return err
	}
	return nil
}

func (m *Messaging) ShutDown() error {
	m.logger.Info().Msg("Shutting down gracefully...")
	m.producer.Stop()
	if m.consumer != nil {
		m.consumer.Stop()
	}
	m.wg.Wait()
	return nil
}

// func (m *Messaging) RequeueMessage(message *nsq.Message) {
// 	m.logger.Info().Msg("Requeue message")
// 	if message.Attempts >= 5 {
// 		m.logger.Error().Msg("Max attempts reached, dropping message")
// 		message.Finish()
// 	} else {
// 		backoffDuration := time.Duration(message.Attempts) * time.Second
// 		message.Requeue(backoffDuration)
// 		m.logger.Info().Msg("Message requeued")
// 	}
// }

func (m *Messaging) GetStats() *nsq.ConsumerStats {
	if m.consumer == nil {
		return nil
	}
	return m.consumer.Stats()
}
