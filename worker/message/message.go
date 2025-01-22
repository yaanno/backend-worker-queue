package message

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	handler "github.com/yaanno/worker/handler"
	"github.com/yaanno/worker/model"
)

const (
	backend_channel = "backend_to_worker"
	worker_channel  = "worker_to_backend"
	nsqd_address    = "nsqd:4150"
	channel         = "channel1"
)

type Messaging interface {
	PublishMessage(msg *model.BackendMessage) error
	ConsumeMessage(msg *model.WorkerMessage) error
	Initialize(handler *handler.BackendResponseHandler) error
	ListenForMessages(ctx context.Context) error
	ShutDown() error
}

type MessagingImpl struct {
	producer *nsq.Producer
	consumer *nsq.Consumer
	config   *nsq.Config
	logger   *zerolog.Logger
	wg       sync.WaitGroup
}

func NewMessaging(config *nsq.Config, logger *zerolog.Logger) Messaging {
	return &MessagingImpl{
		config: config,
		logger: logger,
	}
}

// The rest of the Messaging implementation stays the same

func (m *MessagingImpl) Initialize(handler *handler.BackendResponseHandler) error {
	var err error
	if m.logger == nil {
		m.logger = &zerolog.Logger{}
	}
	m.producer, err = nsq.NewProducer(nsqd_address, m.config)
	if err != nil {
		return err
	}
	m.consumer, err = nsq.NewConsumer(backend_channel, channel, m.config)
	if err != nil {
		return err
	}
	handler.Publish = m.PublishMessage
	m.consumer.AddHandler(handler)

	err = m.consumer.ConnectToNSQD(nsqd_address)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to connect to nsqd")
		return err
	}

	return nil
}

func (m *MessagingImpl) PublishMessage(msg *model.BackendMessage) error {
	message, err := json.Marshal(msg)
	m.logger.Info().Msg("Sending message to backend")
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to marshal message")
		return err
	}
	m.logger.Info().Msg("Publishing message to backend")
	err = m.producer.Publish(worker_channel, message)
	if err != nil {
		return err
	}
	return nil
}

func (m *MessagingImpl) ConsumeMessage(msg *model.WorkerMessage) error {
	m.logger.Info().Msg("Consuming message from backend")
	return nil
}

func (m *MessagingImpl) ShutDown() error {
	m.logger.Info().Msg("Shutting down gracefully...")
	m.producer.Stop()
	m.consumer.Stop()
	m.wg.Wait()
	return nil
}

func (m *MessagingImpl) ListenForMessages(ctx context.Context) error {
	m.logger.Info().Msg("Listening for messages...")
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.logger.Debug().Interface("stats", m.consumer.Stats()).Send()
		<-ctx.Done()
		m.consumer.Stop()
		m.logger.Info().Msg("Stopped listening for messages.")
	}()
	return nil
}
