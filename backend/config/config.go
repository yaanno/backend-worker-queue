package config

import (
	"time"

	"github.com/nsqio/go-nsq"
)

type Config struct {
	NSQConfig       *nsq.Config
	NSQDAddress     string
	MessageInterval time.Duration
	Channels        struct {
		ToWorker   string
		FromWorker string
		Name       string
	}
	MetricsPort     string
	RequestTimeout  time.Duration
	RetryLimit      int
}

func NewConfig() *Config {
	channels := struct {
		ToWorker   string
		FromWorker string
		Name       string
	}{
		ToWorker:   "backend_to_worker",
		FromWorker: "worker_to_backend",
		Name:       "channel1",
	}

	return &Config{
		NSQConfig:       nsq.NewConfig(),
		NSQDAddress:     "nsqd:4150",
		MessageInterval: 10 * time.Second,
		Channels:        channels,
		MetricsPort:     "2113", // Different from worker's 2112
		RequestTimeout:  30 * time.Second,
		RetryLimit:      3,
	}
}
