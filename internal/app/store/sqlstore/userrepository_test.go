package sqlstore_test

import (
	"context"
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/store/sqlstore"
)

func TestUserRepository_Create(t *testing.T) {
	s := sqlstore.NewTestStore(t, "users")
	u := createTestUser(t, s, "user@example.org")

	if u.ID == 0 {
		t.Error("expected user ID to be set")
	}
	if u.Password != "" {
		t.Error("expected password to be cleared")
	}
	if u.EncryptedPassword == "" {
		t.Error("expected encrypted password to be set")
	}

}

func TestUserRepository_FindByEmail(t *testing.T) {
	s := sqlstore.NewTestStore(t, "users")
	ctx := context.Background()
	email := "user@example.org"

	// 1. Не найден → nil, nil
	u, err := s.User().FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u != nil {
		t.Errorf("expected nil user, got %v", u)
	}

	// 2. Создаём
	createTestUser(t, s, email)

	// 3. Найден → вернулся пользователь
	u, err = s.User().FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.Email != email {
		t.Errorf("expected email %q, got %q", email, u.Email)
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	s := sqlstore.NewTestStore(t, "users")
	ctx := context.Background()

	// 1. Не найден → nil, nil
	u, err := s.User().FindByID(ctx, 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u != nil {
		t.Errorf("expected nil user, got %v", u)
	}

	// 2. Создаём
	created := createTestUser(t, s, "user@example.org")

	// 3. Найден по ID
	u, err = s.User().FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, u.ID)
	}
}
