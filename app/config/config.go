package config

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Config struct {
	Service ServiceConfig `mapstructure:"service"`
}

func (config *Config) Validate() error {
	err := validation.ValidateStruct(
		config,
		validation.Field(&config.Service),
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
