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

	//	ListenAndServe запускает TCP-сервер. При каждом входящем запросе вызывает s.router.ServeHTTP(w, r)
	//	передаёт роутер в http.Server. http.Server сам вызывает router.ServeHTTP для каждого запроса
	//	s.router - ServeMux с зарегистрированными маршрутами
	return http.ListenAndServe(s.config.BindAddr, s.router)
}

//// configureLogger переопределяет level
//func (s *APIServer) configureLogger() error {
//	level := slog.LevelInfo
//	switch s.config.LogLevel {
//	case "debug":
//		level = slog.LevelDebug
//	case "info":
//		level = slog.LevelInfo
//	case "warn":
//		level = slog.LevelWarn
//	case "error":
//		level = slog.LevelError
//	}
//
//	//	новый логгер с JSON выводом и level
//	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
//	return nil
//}
//
//// configureStore создаёт store.Store - объект с подключением к MySQL и сохраняет в s.store
//func (s *APIServer) configureStore() error {
//	st := store.New(s.config.Store) //	передает store.Config
//	if err := st.Open(); err != nil {
//		return err //	если MySQL не запущен st.Open() вернет ошибку, сервер не запустится
//	}
//	s.store = st
//	return nil
//}

// configureRouter регистрирует ручки
func (s *APIServer) configureRouter() {
	s.router.HandleFunc("GET /health", s.handleHealth())
	s.router.HandleFunc("POST /users", s.handleCreateUser())
	s.router.HandleFunc("GET /users/{id}", s.handleGetUser())
}
