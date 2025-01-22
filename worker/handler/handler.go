package handler

import (
	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	"github.com/yaanno/worker/model"
)

type MessageResponseHandler struct {
	logger  *zerolog.Logger
	Publish func(msg *model.Response) error
}

func NewBackendResponseHandler(logger *zerolog.Logger) *MessageResponseHandler {
	return &MessageResponseHandler{
		logger: logger,
	}
}

// Example handler method that uses the messaging instance
func (h *MessageResponseHandler) HandleMessage(msg *nsq.Message) error {
	h.logger.Info().Msg("Handling message from backend")
	// Respond back to the backend
	responseMsg := &model.Response{
		// Fill in the response message
		Body: "Response from worker",
		ID:   123456,
	}
	return h.Publish(responseMsg)
}
