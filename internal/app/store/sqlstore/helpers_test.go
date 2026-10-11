package sqlstore_test

import (
	"context"
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/model"
	"github.com/akidamura/http-rest-api/internal/app/store/sqlstore"
)

// createTestUser — хелпер, создаёт пользователя с дефолтным паролем.
func createTestUser(t *testing.T, s *sqlstore.Store, email string) *model.User {
	t.Helper()

	u, err := s.User().Create(context.Background(), &model.User{
		Email:    email,
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return u
}
