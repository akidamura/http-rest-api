package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/akidamura/http-rest-api/internal/app/model"
	"github.com/akidamura/http-rest-api/internal/app/store"
)

func TestUserRepository_Create(t *testing.T) {
	s := store.TestStore(t, "users")

	u, err := s.User().Create(context.Background(), &model.User{
		Email:    "user@example.org",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
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
	s := store.TestStore(t, "users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
	_, err = s.User().Create(ctx, &model.User{
		Email:    email,
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

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
	s := store.TestStore(t, "users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Не найден → nil, nil
	u, err := s.User().FindByID(ctx, 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u != nil {
		t.Errorf("expected nil user, got %v", u)
	}

	// 2. Создаём
	created, err := s.User().Create(ctx, &model.User{
		Email:    "user@example.org",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

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
	if u.Email != "user@example.org" {
		t.Errorf("expected email user@example.org, got %q", u.Email)
	}

}
