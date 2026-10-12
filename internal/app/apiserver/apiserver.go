package apiserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// APIServer ...
type APIServer struct {
	config   *Config
	logger   *slog.Logger
	router   *http.ServeMux
	userRepo UserRepository
	pinger   Pinger
}

// New собирает структуру сервера
func New(config *Config, logger *slog.Logger, userRepo UserRepository, pinger Pinger) *APIServer {
	s := &APIServer{
		config:   config,
		logger:   logger,
		router:   http.NewServeMux(),
		userRepo: userRepo,
		pinger:   pinger,
	}

	s.configureRouter()
	return s
}

// Start запускает сервер
func (s *APIServer) Start() error {
	s.logger.Info("Starting API Server", "addr", s.config.BindAddr)

	// Оборачиваем роутер middleware'ами.
	// Порядок: снаружи внутрь — setRequestID → logRequest → router.
	handler := s.setRequestID(s.logRequest(s.router))

	srv := &http.Server{
		Addr:              s.config.BindAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server error", "err", err)
		}
	}()

	// Ждём SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("shutting down server")

	// Останавливаем с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	s.logger.Info("server stopped")
	return nil
}

// configureRouter регистрирует ручки
func (s *APIServer) configureRouter() {
	s.router.HandleFunc("GET /health", s.handleHealth())
	s.router.HandleFunc("POST /users", s.handleCreateUser())
	s.router.HandleFunc("POST /sessions", s.handleCreateSession())
	s.router.HandleFunc("GET /users/{id}", s.handleGetUser())
	s.router.HandleFunc("PUT /users/{id}", s.handleUpdateUser())
	s.router.HandleFunc("DELETE /users/{id}", s.handleDeleteUser())
}
