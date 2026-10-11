package teststore_test

import (
	"context"
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/model"
	"github.com/akidamura/http-rest-api/internal/app/store/teststore"
)

func TestUserRepository_Create(t *testing.T) {
	s := teststore.New()

	u := &model.User{
		Email:    "user@example.org",
		Password: "secret123",
	}

	created, err := s.User().Create(context.Background(), u)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.ID == 0 {
		t.Error("expected ID to be set")
	}
	if created.Password != "" {
		t.Error("expected Password to be cleared")
	}
	if created.EncryptedPassword == "" {
		t.Error("expected EncryptedPassword to be set")
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	s := teststore.New()

	// 1. Не найден → nil, nil
	u, err := s.User().FindByID(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u != nil {
		t.Errorf("expected nil user, got %v", u)
	}

	// 2. Создаём через хелпер
	created := createTestUser(t, s, "user@example.org")

	// 3. Найден по ID
	u, err = s.User().FindByID(context.Background(), created.ID)
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

func TestUserRepository_FindByEmail(t *testing.T) {
	s := teststore.New()
	email := "user@example.org"

	// 1. Не найден → nil, nil
	u, err := s.User().FindByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u != nil {
		t.Errorf("expected nil user, got %v", u)
	}

	// 2. Создаём через хелпер
	createTestUser(t, s, email)

	// 3. Найден → вернулся пользователь
	u, err = s.User().FindByEmail(context.Background(), email)
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

func TestUserRepository_Update(t *testing.T) {
	s := teststore.New()

	// 1. Не найден → nil, nil
	u, err := s.User().Update(context.Background(), &model.User{
		ID:       999,
		Email:    "nobody@example.org",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u != nil {
		t.Errorf("expected nil user, got %v", u)
	}

	// 2. Создаём
	created := createTestUser(t, s, "user@example.org")

	// 3. Обновляем email
	newEmail := "new@example.org"
	created.Email = newEmail
	created.Password = "newsecret123"

	if _, err := s.User().Update(context.Background(), created); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 4. Проверяем через FindByID — данные обновились в "БД"
	got, err := s.User().FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got == nil {
		t.Fatal("expected user, got nil")
	}
	if got.Email != newEmail {
		t.Errorf("expected email %q, got %q", newEmail, got.Email)
	}
	if got.Password != "" {
		t.Error("expected password to be cleared")
	}
	if got.EncryptedPassword == "" {
		t.Error("expected EncryptedPassword to be set")
	}
}

func TestUserRepository_Delete(t *testing.T) {
	s := teststore.New()

	// 1. Удаление несуществующего — не ошибка
	if err := s.User().Delete(context.Background(), 999); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 2. Создаём
	created := createTestUser(t, s, "user@example.org")

	// 3. Удаляем
	if err := s.User().Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 4. Проверяем, что пользователя больше нет
	got, err := s.User().FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != nil {
		t.Errorf("expected nil user after delete, got %v", got)
	}
}

// createTestUser — хелпер, создаёт пользователя с дефолтным паролем.
func createTestUser(t *testing.T, s *teststore.Store, email string) *model.User {
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
