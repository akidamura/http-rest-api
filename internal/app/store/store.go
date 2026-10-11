package store

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// Store ...
type Store struct {
	config         *Config
	db             *sql.DB
	userRepository *UserRepository
}

// Ping проверяет соединение с БД.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// New создает объект Store с конфигом
func New(config *Config) *Store {
	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	return &Store{config: config}
}

// Open подключается к MySQL через DSN
func (s *Store) Open() error {
	db, err := sql.Open("mysql", s.config.DatabaseURL) // ленивое подключение
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	} // реальная проверка на подключение
	s.db = db
	return nil
}

// Close закрывает подключение к MySQL
func (s *Store) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Store) User() *UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}
	s.userRepository = &UserRepository{
		store: s,
	}
	return s.userRepository
}
