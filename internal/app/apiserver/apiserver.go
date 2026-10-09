package apiserver

import (
	"io"
	"log/slog"
	"net/http"
	"os"
)

// APIServer ...
type APIServer struct {
	config *Config
	logger *slog.Logger
	router *http.ServeMux
}

// New ...
func New(config *Config) *APIServer {
	return &APIServer{
		config: config,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		router: http.NewServeMux(),
	}
}

// Start ...
func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	s.configureRouter()

	s.logger.Info("Starting API Server")

	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *APIServer) configureLogger() error {
	level := slog.LevelInfo
	switch s.config.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	return nil
}

func (s *APIServer) configureRouter() {
	s.router.HandleFunc("GET /hello", s.handleHello())
}

func (s *APIServer) handleHello() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	}
}
