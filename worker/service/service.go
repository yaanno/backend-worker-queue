package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/panjf2000/ants/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/sony/gobreaker"
	"github.com/yaanno/worker/message"
	"github.com/yaanno/worker/metrics"
	"github.com/yaanno/worker/model"
	"golang.org/x/time/rate"
)

type APIConfig struct {
	RateLimit int
	BaseURL   string
}

type WorkerService struct {
	pool        *ants.Pool
	logger      *zerolog.Logger
	httpClient  *http.Client
	apiConfig   APIConfig
	retryLimit  int
	rateLimiter *rate.Limiter
	breaker     *gobreaker.CircuitBreaker
}

func NewWorkerService(
	pool *ants.Pool,
	logger *zerolog.Logger,
	httpClient *http.Client,
	config APIConfig,
	retryLimit int,
) *WorkerService {
	// Create rate limiter based on config
	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), config.RateLimit)

	// Configure circuit breaker
	breakerSettings := gobreaker.Settings{
		Name:        "HTTP",
		MaxRequests: uint32(config.RateLimit),
		Interval:    time.Second * 30,
		Timeout:     time.Second * 60,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.7
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Info().
				Str("from", from.String()).
				Str("to", to.String()).
				Msg("Circuit breaker state changed")
		},
	}

	ws := &WorkerService{
		pool:        pool,
		logger:      logger,
		httpClient:  httpClient,
		apiConfig:   config,
		retryLimit:  retryLimit,
		rateLimiter: limiter,
		breaker:     gobreaker.NewCircuitBreaker(breakerSettings),
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
	timer := prometheus.NewTimer(metrics.MessageProcessingDuration.WithLabelValues("processing"))
	defer timer.ObserveDuration()

	// Update worker pool utilization
	metrics.WorkerUtilization.WithLabelValues("worker").Set(float64(ws.pool.Running()))

	ws.logger.Info().Msg("Processing message")

	if ws.pool == nil {
		ws.logger.Error().Msg("Worker pool is not initialized")
		return errors.New("worker pool not initialized")
	}

	// Validate message
	if len(message.Body) == 0 {
		ws.logger.Error().Msg("Empty message body")
		return errors.New("empty message body")
	}

	// Parse and validate message
	var msg model.Message
	if err := json.Unmarshal(message.Body, &msg); err != nil {
		ws.logger.Error().Err(err).Msg("Failed to parse message")
		return err
	}

	// Submit the task to the ants pool
	err := ws.pool.Submit(func() {
		ws.logger.Info().Msgf("Worker started processing message ID: %s", message.ID)

		retries := 0
		backoff := time.Second

		for {
			// Check rate limit
			if err := ws.rateLimiter.Wait(ctx); err != nil {
				ws.logger.Error().Err(err).Msg("Rate limit exceeded")
				time.Sleep(backoff)
				continue
			}

			// Use circuit breaker for the request
			resp, err := ws.breaker.Execute(func() (interface{}, error) {
				return ws.sendRequest(ctx)
			})

			if err != nil {
				if retries >= ws.retryLimit {
					ws.logger.Error().Err(err).Msg("Max retries reached")
					ws.requeueMessage(message)
					return
				}

				ws.logger.Warn().Err(err).Int("retry", retries).Msg("Retrying request")
				retries++
				backoff *= 2 // Exponential backoff
				time.Sleep(backoff)
				continue
			}

			httpResp := resp.(*http.Response)
			defer httpResp.Body.Close()

			if httpResp.StatusCode >= 500 {
				ws.logger.Error().Int("status", httpResp.StatusCode).Msg("Server error")
				ws.requeueMessage(message)
				return
			}

			// Create response message
			response := &model.Response{
				ID:   msg.ID,
				Body: fmt.Sprintf("Processed message %d with status %d", msg.ID, httpResp.StatusCode),
			}

			// Send response back to backend
			if err := messaging.PublishMessage(response); err != nil {
				ws.logger.Error().Err(err).Msg("Failed to publish response to backend")
				return
			}

			ws.logger.Info().Msgf("Successfully processed message ID: %s", message.ID)
			return
		}
	})

	if err != nil {
		metrics.MessageProcessingTotal.WithLabelValues("error").Inc()
		ws.logger.Error().Err(err).Msg("Failed to submit task to worker pool")
		return err
	}

	metrics.MessageProcessingTotal.WithLabelValues("success").Inc()
	return nil
}

func (ws *WorkerService) sendRequest(ctx context.Context) (*http.Response, error) {
	timer := prometheus.NewTimer(metrics.APIRequestDuration.WithLabelValues("request"))
	defer timer.ObserveDuration()

	// Update circuit breaker metrics
	switch ws.breaker.State() {
	case gobreaker.StateOpen:
		metrics.CircuitBreakerState.Set(0)
	case gobreaker.StateHalfOpen:
		metrics.CircuitBreakerState.Set(1)
	case gobreaker.StateClosed:
		metrics.CircuitBreakerState.Set(2)
	}

	// Execute request through circuit breaker
	resp, err := ws.breaker.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ws.apiConfig.BaseURL, nil)
		if err != nil {
			ws.logger.Error().Err(err).Msg("Error creating request")
			return nil, err
		}

		resp, err := ws.httpClient.Do(req)
		if err != nil {
			ws.logger.Error().Err(err).Msg("Error sending request")
			return nil, err
		}

		// Check if the response status code indicates an error
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			return nil, fmt.Errorf("server error: %s", resp.Status)
		}

		ws.logger.Info().Msgf("Response status: %s, length: %d", resp.Status, resp.ContentLength)
		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	return resp.(*http.Response), nil
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
