package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nsqio/go-nsq"
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
	retryLimit int
}

func NewWorkerService(
	workerPool *pool.WorkerPool,
	messaging *message.Messaging,
	logger *zerolog.Logger,
	httpClient *http.Client,
	retryLimit int,
) *WorkerService {
	return &WorkerService{
		workerPool: workerPool,
		messaging:  messaging,
		logger:     logger,
		httpClient: httpClient,
		retryLimit: retryLimit,
	}
}

func (ws *WorkerService) Start() {
	ws.workerPool.Start()
}

func (ws *WorkerService) Stop() {
	ws.workerPool.Shutdown()
}

func (ws *WorkerService) ProcessMessage(ctx context.Context, message *nsq.Message) error {
	ws.apiUrl = "https://origo.hu"
	// ws.apiUrl = "https://potterapi-fedeperin.vercel.app/es/characters?search=Weasley1"
	ws.logger.Info().Msg("Processing message")
	// Submit the task to the worker pool
	ws.workerPool.Submit(func(ctx context.Context) error {
		ws.logger.Info().Msgf("Worker started processing message ID: %s", message.ID)

		select {
		case <-ctx.Done():
			ws.logger.Error().Msg("Task canceled due to timeout")
			ws.requeueMessage(message)
			return ctx.Err()
		default:
		}

		var msg model.Message
		err := json.Unmarshal(message.Body, &msg)
		if err != nil {
			ws.logger.Error().Err(err).Msg("Failed to unmarshal message")
			ws.requeueMessage(message)
			return err
		}

		startTime := time.Now()
		_, err = ws.sendRequest(ctx)
		elapsedTime := time.Since(startTime)
		ws.logger.Info().Msgf("Task completed in %s", elapsedTime)
		if err != nil {
			ws.logger.Error().Err(err).Msg("Error sending request")
			ws.requeueMessage(message)
			return err
		}

		// Create a response message
		responseMsg := &model.Response{
			ID:   msg.ID,
			Body: msg.Body,
		}
		// Publish the response message
		if err := ws.messaging.PublishMessage(responseMsg); err != nil {
			ws.logger.Error().Err(err).Msg("Error publishing message")
			ws.requeueMessage(message)
			return err
		}
		message.Finish()
		ws.logger.Info().Msgf("Worker finished processing message ID: %s", message.ID)
		return nil
	})

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
	ws.logger.Info().Msgf("Response status: %s, length: %d", resp.Status, resp.ContentLength)
	return resp, nil
}

func (ws *WorkerService) requeueMessage(message *nsq.Message) {
	if message.Attempts >= uint16(ws.retryLimit) {
		ws.logger.Error().Msg("Exceeded retry limit, moving to dead letter queue")
		// Implement dead letter queue or log the message for manual inspection
		message.Finish()
	} else {
		backoffDuration := time.Duration(message.Attempts) * time.Second
		ws.logger.Warn().Msgf("Requeueing message with backoff: %s", backoffDuration)
		message.RequeueWithoutBackoff(backoffDuration)
	}
}

// Helpers

func (ws *WorkerService) GetWorkers() []*pool.Worker {
	return ws.workerPool.GetWorkers()
}

func (ws *WorkerService) GetWorker(id int) *pool.Worker {
	return ws.workerPool.GetWorker(id)
}

func (ws *WorkerService) GetTaskCount() int {
	return ws.workerPool.GetTaskCount()
}

func (ws *WorkerService) GetPoolSize() int {
	return ws.workerPool.GetPoolSize()
}

func (ws *WorkerService) Monitor() []*pool.Worker {
	return ws.GetWorkers()
}

func (ws *WorkerService) PoolSize() int {
	return ws.workerPool.GetPoolSize()
}
