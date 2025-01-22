package handler

import (
	"encoding/json"

	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
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
	var msg model.Message
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal message")
		return err
	}
	h.logger.Info().Interface("Received message from worker:", &msg).Send()
	return nil
}
