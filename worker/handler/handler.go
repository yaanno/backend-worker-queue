package handler

import (
	"context"

	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
)

type WorkerServiceInterface interface {
	ProcessMessage(ctx context.Context, message *nsq.Message) error
}

type MessageResponseHandler struct {
	logger  *zerolog.Logger
	service WorkerServiceInterface
}

func NewMessageResponseHandler(logger *zerolog.Logger, service WorkerServiceInterface) *MessageResponseHandler {
	return &MessageResponseHandler{
		logger:  logger,
		service: service,
	}
}

// HandleMessage method that uses the worker service to process the message
func (h *MessageResponseHandler) HandleMessage(msg *nsq.Message) error {
	h.logger.Info().Msg("Handling message from backend")

	// // Deserialize the message into the custom model
	// var message model.Message
	// err := json.Unmarshal(msg.Body, &message)
	// if err != nil {
	// 	h.logger.Error().Err(err).Msg("Failed to unmarshal message")
	// 	return err
	// }

	// Delegate the processing to the worker service
	err := h.service.ProcessMessage(context.Background(), msg)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to process message")
		return err
	}

	return nil
}
