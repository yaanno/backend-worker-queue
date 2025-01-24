package circuitbreaker

import (
	"time"

	"github.com/sony/gobreaker"
)

// NSQCircuitBreaker wraps the circuit breaker for NSQ operations
type NSQCircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// NewNSQCircuitBreaker creates a new circuit breaker for NSQ operations
func NewNSQCircuitBreaker() *NSQCircuitBreaker {
	settings := gobreaker.Settings{
		Name:        "nsq-circuit-breaker",
		MaxRequests: 3,                // number of requests allowed to pass through when the CircuitBreaker is half-open
		Interval:    10 * time.Second, // cyclic period of the closed state
		Timeout:     60 * time.Second, // period of the open state
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// You could add metrics here to track circuit breaker state changes
		},
	}

	return &NSQCircuitBreaker{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute executes the given request and returns its result.
func (n *NSQCircuitBreaker) Execute(req func() error) error {
	_, err := n.cb.Execute(func() (interface{}, error) {
		return nil, req()
	})
	return err
}
