package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog"
	"github.com/yaanno/worker/message"
	"github.com/yaanno/worker/model"
)

type WorkerService struct {
	pool       *ants.Pool
	logger     *zerolog.Logger
	httpClient *http.Client
	apiUrl     string
	retryLimit int
}

func NewWorkerService(
	pool *ants.Pool,
	logger *zerolog.Logger,
	httpClient *http.Client,
	retryLimit int,
) *WorkerService {
	ws := &WorkerService{
		pool:       pool,
		logger:     logger,
		httpClient: httpClient,
		apiUrl:     "https://origo.hu", // Replace with your default API URL
		retryLimit: retryLimit,
	}
	return ws
}

func (ws *WorkerService) Start() {
	// No explicit start needed for ants pool, it starts automatically
}

func (ws *WorkerService) Stop() {
	// Gracefully shutdown the ants pool
	ws.pool.Release()
}

func (ws *WorkerService) ProcessMessage(ctx context.Context, message *nsq.Message, messaging *message.Messaging) error {
	ws.logger.Info().Msg("Processing message")

	if ws.pool == nil {
		ws.logger.Error().Msg("Worker pool is not initialized")
		return nil
	}

	// Submit the task to the ants pool
	err := ws.pool.Submit(func() {
		ws.logger.Info().Msgf("Worker started processing message ID: %s", message.ID)

		retries := 0

		for {
			// Create a context with timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second) // Adjust timeout as needed
			defer cancel()

			select {
			case <-timeoutCtx.Done():
				ws.logger.Error().Msg("Task canceled due to timeout")
				ws.requeueMessage(message)
				return
			default:
			}

			var msg model.Message
			err := json.Unmarshal(message.Body, &msg)
			if err != nil {
				ws.logger.Error().Err(err).Msg("Failed to unmarshal message")
				ws.requeueMessage(message)
				continue // Continue the retry loop
			}

			startTime := time.Now()
			response, err := ws.sendRequest(timeoutCtx) // Use timeoutCtx for the request
			elapsedTime := time.Since(startTime)
			ws.logger.Info().Msgf("Task completed in %s", elapsedTime)
			if err != nil || response.StatusCode != http.StatusOK {
				retries++
				if retries >= ws.retryLimit {
					ws.logger.Error().Err(err).Msgf("Error sending request (attempt %d)", retries)
					ws.requeueMessage(message)
					return // Exit the worker function
				}
				ws.logger.Warn().Err(err).Msgf("Retrying request (attempt %d)", retries)
				time.Sleep(time.Duration(retries) * time.Second) // Exponential backoff
				continue                                         // Continue the retry loop
			}

			// Create a response message
			responseMsg := &model.Response{
				ID:   msg.ID,
				Body: msg.Body,
			}
			// Publish the response message
			if err := messaging.PublishMessage(responseMsg); err != nil {
				ws.logger.Error().Err(err).Msg("Error publishing message")
				ws.requeueMessage(message)
				continue // Continue the retry loop
			}
			message.Finish()
			ws.logger.Info().Msgf("Worker finished processing message ID: %s", message.ID)
			return // Exit the worker function successfully
		}
	})

	if err != nil {
		ws.logger.Error().Err(err).Msg("Error submitting task to pool")
		return err
	}

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
		message.Requeue(backoffDuration)
	}
}

func (ws *WorkerService) Monitor() {
	ws.logger.Info().Msg("Monitoring worker pool")
	ws.logger.Info().Msgf("Pool size: %d", ws.pool.Cap())
	ws.logger.Info().Msgf("Running workers: %d", ws.pool.Running())
	ws.logger.Info().Msgf("Free workers: %d", ws.pool.Free())
	ws.logger.Info().Msgf("Waiting tasks: %d", ws.pool.Waiting())
}
