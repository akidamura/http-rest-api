package apiserver

import (
	"log/slog"
	"net/http"
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

	// ListenAndServe запускает TCP-сервер.
	// При каждом запросе вызывает handler.ServeHTTP(w, r),
	// где handler — цепочка middleware + router.
	return http.ListenAndServe(s.config.BindAddr, handler)
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
