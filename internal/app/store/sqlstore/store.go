package sqlstore

import (
	"context"
	"database/sql"
)

// Store ...
type Store struct {
	db             *sql.DB
	userRepository *UserRepository
}

// New создает объект Store с конфигом
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// Ping проверяет соединение с БД.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
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
