package message

import (
	"encoding/json"

	"github.com/nsqio/go-nsq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/yaanno/backend/config"
	handler "github.com/yaanno/backend/handler"
	"github.com/yaanno/backend/metrics"
	"github.com/yaanno/backend/model"
)

type Messaging interface {
	PublishMessage(msg *model.Response) error
	ConsumeMessage(msg *model.Message) error
	Initialize(handler *handler.MessageResponseHandler) error
	ShutDown() error
}

type MessagingImpl struct {
	producer *nsq.Producer
	consumer *nsq.Consumer
	config   *config.Config
	logger   *zerolog.Logger
}

func NewMessaging(config *config.Config, logger *zerolog.Logger) Messaging {
	return &MessagingImpl{
		config: config,
		logger: logger,
	}
}

func (m *MessagingImpl) Initialize(handler *handler.MessageResponseHandler) error {
	var err error
	if m.logger == nil {
		m.logger = &zerolog.Logger{}
	}

	// Initialize producer
	m.producer, err = nsq.NewProducer(m.config.NSQDAddress, m.config.NSQConfig)
	if err != nil {
		metrics.NSQConnectionStatus.Set(0)
		return err
	}

	// Initialize consumer
	m.consumer, err = nsq.NewConsumer(m.config.Channels.FromWorker, m.config.Channels.Name, m.config.NSQConfig)
	if err != nil {
		metrics.NSQConnectionStatus.Set(0)
		return err
	}

	m.consumer.AddHandler(handler)
	err = m.consumer.ConnectToNSQD(m.config.NSQDAddress)
	if err != nil {
		metrics.NSQConnectionStatus.Set(0)
		return err
	}

	metrics.NSQConnectionStatus.Set(1)
	return nil
}

func (m *MessagingImpl) PublishMessage(msg *model.Response) error {
	timer := prometheus.NewTimer(metrics.MessageProcessingDuration.WithLabelValues("send"))
	defer timer.ObserveDuration()

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		metrics.MessagesSent.WithLabelValues("error").Inc()
		return err
	}

	err = m.producer.Publish(m.config.Channels.ToWorker, jsonMsg)
	if err != nil {
		metrics.MessagesSent.WithLabelValues("error").Inc()
		return err
	}

	metrics.MessagesSent.WithLabelValues("success").Inc()
	metrics.MessageQueueDepth.Inc()
	return nil
}

func (m *MessagingImpl) ConsumeMessage(msg *model.Message) error {
	m.logger.Info().Msg("Consuming message from worker")
	return nil
}

func (m *MessagingImpl) ShutDown() error {
	if m.consumer != nil {
		m.consumer.Stop()
	}
	if m.producer != nil {
		m.producer.Stop()
	}
	metrics.NSQConnectionStatus.Set(0)
	return nil
}
