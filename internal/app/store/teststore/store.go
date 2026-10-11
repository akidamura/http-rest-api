package teststore

import (
	"context"

	"github.com/akidamura/http-rest-api/internal/app/model"
)

// Store ...
type Store struct {
	userRepository *UserRepository
}

// New ...
func New() *Store {
	return &Store{}
}

// User ...
func (s *Store) User() *UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}
	s.userRepository = &UserRepository{
		store:  s,
		users:  make(map[int]*model.User),
		nextID: 0,
	}
	return s.userRepository
}

// Ping всегда возвращает nil — teststore всегда "доступен".
func (s *Store) Ping(ctx context.Context) error {
	return nil
}
