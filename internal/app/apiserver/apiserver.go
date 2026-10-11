package apiserver

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/akidamura/http-rest-api/internal/app/store"
)

// APIServer ...
type APIServer struct {
	config *Config
	logger *slog.Logger
	router *http.ServeMux
	store  *store.Store
}

// New собирает структуру сервера
func New(config *Config) *APIServer {
	return &APIServer{
		config: config,                                        //	содержит весь конфиг, включая config.Store
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)), //	логгер по умолчанию
		router: http.NewServeMux(),                            //	пустой ServeMux
		//	store не указан → nil
	}
}

// Start подключает сервер
func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	if err := s.configureStore(); err != nil {
		return err
	}
	defer s.store.Close()

	s.configureRouter() //	ручки используют s.store, поэтому store уже должен быть

	s.logger.Info("Starting API Server")

	//	ListenAndServe запускает TCP-сервер. При каждом входящем запросе вызывает s.router.ServeHTTP(w, r)
	//	передаёт роутер в http.Server. http.Server сам вызывает router.ServeHTTP для каждого запроса
	//	s.router - ServeMux с зарегистрированными маршрутами
	return http.ListenAndServe(s.config.BindAddr, s.router)
}

// configureLogger переопределяет level
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

	//	новый логгер с JSON выводом и level
	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	return nil
}

// configureStore создаёт store.Store - объект с подключением к MySQL и сохраняет в s.store
func (s *APIServer) configureStore() error {
	st := store.New(s.config.Store) //	передает store.Config
	if err := st.Open(); err != nil {
		return err //	если MySQL не запущен st.Open() вернет ошибку, сервер не запустится
	}
	s.store = st
	return nil
}

// configureRouter регистрирует ручки
func (s *APIServer) configureRouter() {
	s.router.HandleFunc("GET /health", s.handleHealth())
	s.router.HandleFunc("POST /users", s.handleCreateUser())
	s.router.HandleFunc("GET /users/{id}", s.handleGetUser())
}
