package service

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/yaanno/worker/message"
	"github.com/yaanno/worker/model"
	"github.com/yaanno/worker/pool"
)

type WorkerService struct {
	workerPool *pool.WorkerPool
	messaging  *message.Messaging
	logger     *zerolog.Logger
	httpClient *http.Client
	apiUrl     string
}

func NewWorkerService(workerPool *pool.WorkerPool, messaging *message.Messaging, logger *zerolog.Logger, httpClient *http.Client) *WorkerService {
	return &WorkerService{
		workerPool: workerPool,
		messaging:  messaging,
		logger:     logger,
		httpClient: httpClient,
	}
}

func (ws *WorkerService) Start() {
	ws.workerPool.Start()
}

func (ws *WorkerService) Stop() {
	ws.workerPool.Stop()
}

func (ws *WorkerService) ProcessMessage(ctx context.Context, message *model.Message) error {
	ws.apiUrl = "https://google.com"
	ws.logger.Info().Msg("Processing message")
	ws.workerPool.Submit(func(ctx context.Context) {
		_, err := ws.sendRequest(ctx)
		if err != nil {
			ws.logger.Error().Err(err).Msg("Error sending request")
			return
		}
		responseMsg := &model.Response{
			ID: message.ID,
			// Body: response.Body,
			Body: "payload",
		}
		if err := ws.messaging.PublishMessage(responseMsg); err != nil {
			ws.logger.Error().Err(err).Msg("Error publishing message")
			return
		}
	})
	ws.logger.Info().Msg("Processing message")
	return nil
}

func (ws *WorkerService) sendRequest(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ws.apiUrl, nil)
	if err != nil {
		ws.logger.Error().Err(err).Msg("Error creating request")
		return nil, err
	}

	resp, err := ws.httpClient.Do(req)
	if err != nil {
		ws.logger.Error().Err(err).Msg("Error sending request")
		return nil, err
	}
	defer resp.Body.Close()
	ws.logger.Info().Msgf("Response status: %s", resp.Status)
	return resp, nil
}
