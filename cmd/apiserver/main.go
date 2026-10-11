package main

import (
	"log/slog"
	"os"

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

	st := store.New(cfg.Store)
	if err := st.Open(); err != nil {
		logger.Error("failed to open store", "err", err)
		os.Exit(1)
	}
	defer st.Close()

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
