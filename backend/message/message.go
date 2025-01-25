package message

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nsqio/go-nsq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/yaanno/backend/circuitbreaker"
	"github.com/yaanno/backend/config"
	handler "github.com/yaanno/backend/handler"
	"github.com/yaanno/backend/metrics"
	"github.com/yaanno/backend/model"
	"github.com/yaanno/backend/retry"
)

type Messaging interface {
	PublishMessage(ctx context.Context, msg *model.Response) error
	ConsumeMessage(ctx context.Context, msg *model.Message) error
	Initialize(ctx context.Context, handler *handler.MessageResponseHandler) error
	ShutDown(ctx context.Context) error
}

type MessagingImpl struct {
	producer        *nsq.Producer
	consumer        *nsq.Consumer
	config          *config.Config
	logger          *zerolog.Logger
	mu              sync.RWMutex
	producerBreaker *circuitbreaker.NSQCircuitBreaker
	consumerBreaker *circuitbreaker.NSQCircuitBreaker
	retryConfig     *retry.RetryConfig
}

func NewMessaging(config *config.Config, logger *zerolog.Logger) Messaging {
	return &MessagingImpl{
		config:          config,
		logger:          logger,
		producerBreaker: circuitbreaker.NewNSQCircuitBreaker(),
		consumerBreaker: circuitbreaker.NewNSQCircuitBreaker(),
		retryConfig:     retry.DefaultRetryConfig(),
	}
}

func (m *MessagingImpl) Initialize(ctx context.Context, handler *handler.MessageResponseHandler) error {
	if handler == nil {
		return ErrNilHandler
	}

	if err := m.validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.logger == nil {
		m.logger = &zerolog.Logger{}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Initialize producer with circuit breaker and retry
		err := retry.DoWithRetry(ctx, "initialize_producer", m.logger, m.retryConfig, func() error {
			return m.producerBreaker.Execute(func() error {
				producer, err := nsq.NewProducer(m.config.NSQ.Address, nsq.NewConfig())
				if err != nil {
					metrics.NSQConnectionStatus.WithLabelValues("producer").Set(0)
					return fmt.Errorf("create producer: %w", err)
				}
				m.producer = producer
				return nil
			})
		})
		if err != nil {
			return fmt.Errorf("failed to initialize producer: %w", err)
		}

		// Initialize consumer with circuit breaker and retry
		err = retry.DoWithRetry(ctx, "initialize_consumer", m.logger, m.retryConfig, func() error {
			return m.consumerBreaker.Execute(func() error {
				consumer, err := nsq.NewConsumer(m.config.NSQ.Channels.FromWorker, m.config.NSQ.Channels.Name, nsq.NewConfig())
				if err != nil {
					metrics.NSQConnectionStatus.WithLabelValues("consumer").Set(0)
					return fmt.Errorf("create consumer: %w", err)
				}
				m.consumer = consumer

				// Wrap the handler with circuit breaker and retry
				m.consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
					return retry.DoWithRetry(ctx, "handle_message", m.logger, m.retryConfig, func() error {
						return m.consumerBreaker.Execute(func() error {
							return handler.HandleMessage(message)
						})
					})
				}))

				err = m.consumer.ConnectToNSQD(m.config.NSQ.Address)
				if err != nil {
					metrics.NSQConnectionStatus.WithLabelValues("connection").Set(0)
					return fmt.Errorf("connect to NSQD: %w", err)
				}
				return nil
			})
		})
		if err != nil {
			return fmt.Errorf("failed to initialize consumer: %w", err)
		}

		metrics.NSQConnectionStatus.WithLabelValues("producer").Set(1)
		m.logger.Info().
			Str("nsqd_address", m.config.NSQ.Address).
			Str("channel", m.config.NSQ.Channels.Name).
			Msg("Successfully initialized NSQ messaging")
		return nil
	}
}

func (m *MessagingImpl) PublishMessage(ctx context.Context, msg *model.Response) error {
	if msg == nil {
		return ErrNilMessage
	}

	timer := prometheus.NewTimer(metrics.MessageProcessingDuration.WithLabelValues("send"))
	defer timer.ObserveDuration()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.producer == nil {
		return ErrProducerNotInitialized
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			metrics.MessagesSent.WithLabelValues("error").Inc()
			return fmt.Errorf("marshal message: %w", err)
		}

		msgSize := float64(len(jsonMsg))
		// metrics.MessageSize.WithLabelValues("publish").Observe(msgSize)

		// Use circuit breaker and retry for publish operation
		err = retry.DoWithRetry(ctx, "publish_message", m.logger, m.retryConfig, func() error {
			return m.producerBreaker.Execute(func() error {
				return m.producer.Publish(m.config.NSQ.Channels.ToWorker, jsonMsg)
			})
		})

		if err != nil {
			metrics.MessagesSent.WithLabelValues("error").Inc()
			return fmt.Errorf("publish message: %w", err)
		}

		metrics.MessagesSent.WithLabelValues("success").Inc()

		m.logger.Debug().
			Int("msg_id", msg.ID).
			Float64("size_bytes", msgSize).
			Msg("Successfully published message")
		return nil
	}
}

func (m *MessagingImpl) ConsumeMessage(ctx context.Context, msg *model.Message) error {
	if msg == nil {
		return ErrNilMessage
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.consumer == nil {
		return ErrConsumerNotInitialized
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Use circuit breaker and retry for consume operation
		err := retry.DoWithRetry(ctx, "consume_message", m.logger, m.retryConfig, func() error {
			return m.consumerBreaker.Execute(func() error {
				m.logger.Info().
					Int("msg_id", msg.ID).
					Msg("Consuming message from worker")
				return nil
			})
		})

		if err != nil {
			metrics.MessagesReceived.WithLabelValues("error").Inc()
			return fmt.Errorf("consume message: %w", err)
		}

		metrics.MessagesReceived.WithLabelValues("success").Inc()
		return nil
	}
}

func (m *MessagingImpl) ShutDown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if m.consumer != nil {
			m.consumer.Stop()
			m.logger.Info().Msg("NSQ consumer stopped")
		}
		if m.producer != nil {
			m.producer.Stop()
			m.logger.Info().Msg("NSQ producer stopped")
			metrics.NSQConnectionStatus.WithLabelValues("producer").Set(0)
		}
		metrics.NSQConnectionStatus.WithLabelValues("connection").Set(0)
		return nil
	}
}
