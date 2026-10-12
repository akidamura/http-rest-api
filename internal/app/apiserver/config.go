package apiserver

import (
	"github.com/akidamura/http-rest-api/internal/config"
)

// Config apiserver
type Config struct {
	BindAddr   string
	LogLevel   string
	LogFormat  string
	SessionKey string
}

// NewConfig собирает конфиг apiserver из ENV
func NewConfig() *Config {
	return &Config{
		BindAddr:   config.GetEnv("BIND_ADDR", ":8080"),
		LogLevel:   config.GetEnv("LOG_LEVEL", "info"),
		LogFormat:  config.GetEnv("LOG_FORMAT", "text"),
		SessionKey: config.GetEnv("SESSION_KEY", "default-secret-key-change-in-prod"),
	}
}
