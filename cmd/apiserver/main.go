package main

import (
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/akidamura/http-rest-api/internal/app/store"
	"github.com/akidamura/http-rest-api/internal/app/store/sqlstore"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"

	"github.com/akidamura/http-rest-api/internal/app/apiserver"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := apiserver.NewConfig()

	logger := newLogger(cfg.LogLevel, cfg.LogFormat)

	// 1. Открываем соединение
	db, err := sql.Open("mysql", store.BuildDSN())
	if err != nil {
		logger.Error("failed to open db", "err", err)
		os.Exit(1)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		logger.Error("Error pinging database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// 2. Собираем Store поверх готового *sql.DB
	st := sqlstore.New(db)

	// 3. SessionStore
	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionKey))

	// 4. Собираем сервер
	s := apiserver.New(cfg, logger, st.User(), st, sessionStore)
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

// Переключатель формата вывода для логгера, в .ENV
func newLogger(level string, format string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(level),
	}

	// Для text - сокращенный формат времени
	if format != "json" {
		opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String(a.Key, a.Value.Time().Format("15:04:05.000"))
			}
			return a
		}
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
