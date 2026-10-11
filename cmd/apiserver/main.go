package main

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/akidamura/http-rest-api/internal/app/apiserver"
	"github.com/akidamura/http-rest-api/internal/app/store"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := apiserver.NewConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))

	// 1. Открываем соединение
	db, err := sql.Open("mysql", store.BuildDSN())
	if err != nil {
		logger.Error("failed to open db", "err", err)
		return
	}
	if err := db.Ping(); err != nil {
		logger.Error("Error pinging database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// 2. Собираем Store поверх готового *sql.DB
	st := store.New(db)

	// 3. Собираем сервер
	s := apiserver.New(cfg, logger, st.User(), st)
	if err := s.Start(); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
