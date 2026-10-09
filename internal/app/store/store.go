package store

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// Store ...
type Store struct {
	config *Config
	db     *sql.DB
}

// New ...
func New(config *Config) *Store {
	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	return &Store{config: config}
}

// Open ...
func (s *Store) Open() error {
	db, err := sql.Open("mysql", s.config.DatabaseURL)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	s.db = db
	return nil
}

// Close ...
func (s *Store) Close() {
	if s.db != nil {
		s.db.Close()
	}
}
