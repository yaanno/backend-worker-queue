package handler

import (
	"encoding/json"

	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	"github.com/yaanno/backend/model"
)

type WorkerResponseHandler struct {
	logger *zerolog.Logger
}

func NewWorkerResponseHandler(logger *zerolog.Logger) *WorkerResponseHandler {
	return &WorkerResponseHandler{
		logger: logger,
	}
}

func (h *WorkerResponseHandler) HandleMessage(m *nsq.Message) error {
	var msg model.WorkerMessage
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal message")
		return err
	}
	h.logger.Info().Interface("Received message from worker:", &msg).Send()
	return nil
}
