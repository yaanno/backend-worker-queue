package message

import (
	"errors"
	"fmt"
)

var (
	// ErrNilMessage is returned when attempting to publish a nil message
	ErrNilMessage = errors.New("message cannot be nil")
	// ErrNilHandler is returned when attempting to initialize with a nil handler
	ErrNilHandler = errors.New("message handler cannot be nil")
	// ErrProducerNotInitialized is returned when attempting to publish before initialization
	ErrProducerNotInitialized = errors.New("nsq producer not initialized")
	// ErrConsumerNotInitialized is returned when attempting to consume before initialization
	ErrConsumerNotInitialized = errors.New("nsq consumer not initialized")
	// ErrInvalidConfig is returned when configuration validation fails
	ErrInvalidConfig = errors.New("invalid configuration")
)

func (m *MessagingImpl) validateConfig() error {
	if m.config == nil {
		return fmt.Errorf("%w: config is nil", ErrInvalidConfig)
	}
	if m.config.NSQDAddress == "" {
		return fmt.Errorf("%w: NSQDAddress is empty", ErrInvalidConfig)
	}
	if m.config.NSQConfig == nil {
		return fmt.Errorf("%w: NSQConfig is nil", ErrInvalidConfig)
	}
	if m.config.Channels.Name == "" {
		return fmt.Errorf("%w: channel name is empty", ErrInvalidConfig)
	}
	if m.config.Channels.ToWorker == "" {
		return fmt.Errorf("%w: ToWorker channel is empty", ErrInvalidConfig)
	}
	if m.config.Channels.FromWorker == "" {
		return fmt.Errorf("%w: FromWorker channel is empty", ErrInvalidConfig)
	}
	return nil
}
