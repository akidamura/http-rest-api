package apiserver

import (
	"github.com/akidamura/http-rest-api/internal/config"
)

// Config apiserver
type Config struct {
	BindAddr string
	LogLevel string
}

// NewConfig собирает конфиг apiserver из ENV
func NewConfig() *Config {
	return &Config{
		BindAddr: config.GetEnv("BIND_ADDR", ":8080"),
		LogLevel: config.GetEnv("LOG_LEVEL", "info"),
	}
}
