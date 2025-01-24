package handler

import (
	"context"

	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	"github.com/yaanno/worker/message"
	"github.com/yaanno/worker/service"
)

// type WorkerServiceInterface interface {
// 	ProcessMessage(ctx context.Context, message *nsq.Message) error
// }

type MessageResponseHandler struct {
	logger    *zerolog.Logger
	service   *service.WorkerService
	messaging *message.Messaging
}

func NewMessageResponseHandler(ctx context.Context, logger *zerolog.Logger, service *service.WorkerService, messaging *message.Messaging) *MessageResponseHandler {
	return &MessageResponseHandler{
		logger:    logger,
		service:   service,
		messaging: messaging, // Add messaging field
	}
}

func (h *MessageResponseHandler) HandleMessage(msg *nsq.Message) error {
	h.logger.Info().Msg("Handling message from backend")
	err := h.service.ProcessMessage(context.Background(), msg, h.messaging)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to process message")
		return err
	}
	return nil
}
