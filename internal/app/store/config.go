package store

import (
	"fmt"
	"log"

	"github.com/akidamura/http-rest-api/internal/config"
)

// Config ...
type Config struct {
	DatabaseURL string
}

func NewConfig() *Config {
	url := config.GetEnv("DATABASE_URL", "")
	if url == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	return &Config{DatabaseURL: buildDSN()}
}

func buildDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.GetEnv("DB_USER", "root"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "3306"),
		config.GetEnv("DB_NAME", "restapi_dev"),
	)
}
