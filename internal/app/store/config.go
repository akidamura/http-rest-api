package store

import (
	"fmt"
	"log"

	"github.com/akidamura/http-rest-api/internal/config"
)

// Config store
type Config struct {
	DatabaseURL string
}

// NewConfig собирает конфиг store из ENV
func NewConfig() *Config {
	dsn := buildDSN()

	if config.GetEnv("DB_PASSWORD", "") == "" {
		log.Fatal("DB_PASSWORD environment variable not set")
	}
	if config.GetEnv("DB_USER", "") == "" {
		log.Fatal("DB_USER environment variable not set")
	}

	return &Config{DatabaseURL: dsn}
}

func buildDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.GetEnv("DB_USER", "root"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "3306"),
		config.GetEnv("DB_NAME", "dbname"),
	)
}
