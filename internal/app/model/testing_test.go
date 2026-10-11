package model

import "testing"

// NewTestUser создаёт пользователя с дефолтными полями для тестов.
func NewTestUser(t *testing.T) *User {
	t.Helper()
	return &User{
		Email:    "user@example.org",
		Password: "secret123",
	}
}
