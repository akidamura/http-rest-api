package store

import (
	"fmt"

	"github.com/akidamura/http-rest-api/internal/config"
)

// BuildDSN собирает строку подключения к MySQL из ENV.
func BuildDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.GetEnv("DB_USER", "root"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "3306"),
		config.GetEnv("DB_NAME", "dbname"),
	)
}
