package apiserver

import "os"

// Config ...
type Config struct {
	BindAddr string
	LogLevel string
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr: getEnv("BIND_ADDR", ":8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
