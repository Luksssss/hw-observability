package config

import (
	"github.com/joho/godotenv"
	// _ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
)

// Config represents application configuration.
type Config struct {
	// Development mode enables dev-only features
	Development bool `envconfig:"DEVELOPMENT" default:"false"`

	// Port to listen on
	Port     int `envconfig:"SERVER_PORT" default:"8080"`
	PortProm int `envconfig:"PROMETHEUS_PORT" default:"8081"`

	// Config trace
	TrConfType  string  `envconfig:"TRACE_CONF_TYPE" default:"const"`
	TrConfParam float64 `envconfig:"TRACE_CONF_PARAM" default:"1"`

	// Site
	Site string `envconfig:"SITE"`
}

// ReadConfig считываем  env-конфиг
func ReadConfig(cfg *Config) error {
	godotenv.Load(".env")
	err := envconfig.Process("", cfg)
	return err
}
