package config

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Config struct {
	Service  ServiceConfig  `mapstructure:"service"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitMq"`
}

func (config *Config) Validate() error {
	err := validation.ValidateStruct(
		config,
		validation.Field(&config.Service),
		validation.Field(&config.RabbitMQ),
	)
	return err
}

type ServiceConfig struct {
	LogLevel    string `mapstructure:"logLevel"`
	Name        string
	Environment string
	Port        string
}

func (config ServiceConfig) Validate() error {
	err := validation.ValidateStruct(
		&config,
		validation.Field(&config.Port, is.Port),
		validation.Field(&config.LogLevel, validation.Required),
		validation.Field(&config.Environment, validation.Required),
	)
	return err
}

type QueueConfig struct {
	Name          string      `mapstructure:"name"`
	Retry         RetryConfig `mapstructure:"retry"`
	PrefetchCount int         `mapstructure:"prefetchCount"`
}

type RetryConfig struct {
	MaxRetries        int `mapstructure:"maxRetries"`
	RetryDelaySeconds int `mapstructure:"retryDelaySeconds"`
}

type RabbitMQConfig struct {
	Hosts    []string
	Username string
	Password string
	Queues   []QueueConfig `mapstructure:"queues"`
}

func (config RabbitMQConfig) Validate() error {
	err := validation.ValidateStruct(
		&config,
		validation.Field(&config.Hosts, validation.Required),
		validation.Field(&config.Username, validation.Required),
		validation.Field(&config.Password, validation.Required),
		validation.Field(&config.Queues, validation.Required),
	)
	return err
}

func (config *Config) GetQueueConfig(queueName string) *QueueConfig {
	for _, queue := range config.RabbitMQ.Queues {
		if queue.Name == queueName {
			return &queue
		}
	}
	return nil
}
