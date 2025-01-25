package handler

import (
	"encoding/json"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/yaanno/backend/metrics"
	"github.com/yaanno/backend/model"
)

type MessageResponseHandler struct {
	logger *zerolog.Logger
}

func NewMessageResponseHandler(logger *zerolog.Logger) *MessageResponseHandler {
	return &MessageResponseHandler{
		logger: logger,
	}
}

func (h *MessageResponseHandler) HandleMessage(m *nsq.Message) error {
	timer := prometheus.NewTimer(metrics.MessageProcessingDuration.WithLabelValues("receive"))
	defer timer.ObserveDuration()

	var msg model.Message
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal message")
		metrics.MessagesReceived.WithLabelValues("error").Inc()
		// Requeue the message if it's not a permanent error
		if m.Attempts < 3 {
			m.RequeueWithoutBackoff(time.Second * 5)
			return nil
		}
		return err
	}

	h.logger.Info().Interface("message", msg).Msg("Received message from worker")
	metrics.MessagesReceived.WithLabelValues("success").Inc()
	metrics.MessageQueueDepth.WithLabelValues("worker").Dec()
	return nil
}
