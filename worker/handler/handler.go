package handler

import (
	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	"github.com/yaanno/worker/model"
)

type BackendResponseHandler struct {
	logger  *zerolog.Logger
	Publish func(msg *model.BackendMessage) error
}

func NewBackendResponseHandler(logger *zerolog.Logger) *BackendResponseHandler {
	return &BackendResponseHandler{
		logger: logger,
	}
}

// Example handler method that uses the messaging instance
func (h *BackendResponseHandler) HandleMessage(msg *nsq.Message) error {
	h.logger.Info().Msg("Handling message from backend")
	// Respond back to the backend
	responseMsg := &model.BackendMessage{
		// Fill in the response message
		Body: "Response from worker",
		ID:   123456,
	}
	return h.Publish(responseMsg)
}
