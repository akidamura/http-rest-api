package apiserver

import (
	"github.com/akidamura/http-rest-api/internal/app/store"
	"github.com/akidamura/http-rest-api/internal/config"
)

// Config ...
type Config struct {
	BindAddr string
	LogLevel string
	Store    *store.Config
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr: config.GetEnv("BIND_ADDR", ":8080"),
		LogLevel: config.GetEnv("LOG_LEVEL", "info"),
		Store:    store.NewConfig(),
	}
}
