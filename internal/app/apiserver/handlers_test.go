package apiserver

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/model"
	"github.com/gorilla/sessions"
)

// --- фейки ---

type fakeUserRepo struct {
	createFn      func(ctx context.Context, u *model.User) (*model.User, error)
	findByIDFn    func(ctx context.Context, id int) (*model.User, error)
	findByEmailFn func(ctx context.Context, email string) (*model.User, error)
	updateFn      func(ctx context.Context, u *model.User) (*model.User, error)
	deleteFn      func(ctx context.Context, id int) error
}

func (f *fakeUserRepo) Create(ctx context.Context, u *model.User) (*model.User, error) {
	return f.createFn(ctx, u)
}
func (f *fakeUserRepo) FindByID(ctx context.Context, id int) (*model.User, error) {
	return f.findByIDFn(ctx, id)
}
func (f *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return f.findByEmailFn(ctx, email)
}
func (f *fakeUserRepo) Update(ctx context.Context, u *model.User) (*model.User, error) {
	return f.updateFn(ctx, u)
}
func (f *fakeUserRepo) Delete(ctx context.Context, id int) error {
	return f.deleteFn(ctx, id)
}

type fakePinger struct{ err error }

func (f *fakePinger) Ping(ctx context.Context) error { return f.err }

// --- хелперы ---

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		createFn: func(_ context.Context, u *model.User) (*model.User, error) {
			return u, nil
		},
		findByIDFn: func(_ context.Context, _ int) (*model.User, error) {
			return nil, nil
		},
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, nil
		},
		updateFn: func(_ context.Context, u *model.User) (*model.User, error) {
			return u, nil
		},
		deleteFn: func(_ context.Context, _ int) error {
			return nil
		},
	}
}

func newTestAPI(repo UserRepository, p Pinger) *APIServer {
	sessionStore := sessions.NewCookieStore([]byte("test-secret-key"))
	return New(&Config{}, slog.Default(), repo, p, sessionStore)
}

// --- ТЕСТЫ ---

// --- Health ---

// TestHandleHealth_OK БД жива → 200
func TestHandleHealth_OK(t *testing.T) {
	api := newTestAPI(newFakeUserRepo(), &fakePinger{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	api.handleHealth().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestHandleHealth_DBDown БД упала → 503
func TestHandleHealth_DBDown(t *testing.T) {
	api := newTestAPI(newFakeUserRepo(), &fakePinger{err: context.DeadlineExceeded})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	api.handleHealth().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

// --- CreateUser ---

// TestHandleCreateUser_OK Успех → 201
func TestHandleCreateUser_OK(t *testing.T) {
	api := newTestAPI(newFakeUserRepo(), &fakePinger{})

	body := `{"email":"user@example.org","password":"secret123"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	t.Logf("request body: %s", body)

	api.handleCreateUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d, want %d, resp body=%q", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

// TestHandleCreateUser_BadJSON Плохой JSON → 400
func TestHandleCreateUser_BadJSON(t *testing.T) {
	api := newTestAPI(newFakeUserRepo(), &fakePinger{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{not json"))
	req.Header.Set("Content-Type", "application/json")

	api.handleCreateUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleCreateUser_ValidationError Ошибка валидации → 400
func TestHandleCreateUser_ValidationError(t *testing.T) {
	repo := newFakeUserRepo()
	repo.createFn = func(_ context.Context, _ *model.User) (*model.User, error) {
		return nil, model.ErrEmailRequired
	}
	api := newTestAPI(repo, &fakePinger{})

	body := `{"email":"","password":""}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))

	t.Logf("request body: %s", body)

	api.handleCreateUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want %d, resp body=%q", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

// --- CreateSession ---

// TestHandleCreateSession_OK успех → 200
func TestHandleCreateSession_OK(t *testing.T) {
	// 1. Готовим пользователя с хешем
	u := &model.User{Email: "user@example.org", Password: "secret123"}
	if err := u.BeforeCreate(); err != nil {
		t.Fatal(err)
	}

	// 2. Фейк возвращает этого пользователя
	repo := newFakeUserRepo()
	repo.findByEmailFn = func(_ context.Context, _ string) (*model.User, error) {
		return u, nil
	}

	api := newTestAPI(repo, &fakePinger{})

	// 3. Запрос
	body := `{"email":"user@example.org","password":"secret123"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body))

	t.Logf("request body: %s", body)

	// 4. Дергаем хендлер
	api.handleCreateSession().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d, body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
}

// TestHandleCreateSession_UserNotFound пользователь не найден → 401
func TestHandleCreateSession_UserNotFound(t *testing.T) {
	// findByEmailFn дефолтный возвращает nil, nil — уже подходит
	api := newTestAPI(newFakeUserRepo(), &fakePinger{})

	body := `{"email":"nobody@example.org","password":"secret123"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body))

	api.handleCreateSession().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// TestHandleCreateSession_WrongPassword пароль неверный → 401
func TestHandleCreateSession_WrongPassword(t *testing.T) {
	u := &model.User{Email: "user@example.org", Password: "secret123"}
	if err := u.BeforeCreate(); err != nil {
		t.Fatal(err)
	}

	repo := newFakeUserRepo()
	repo.findByEmailFn = func(_ context.Context, _ string) (*model.User, error) {
		return u, nil
	}

	api := newTestAPI(repo, &fakePinger{})

	body := `{"email":"user@example.org","password":"WRONG"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body))

	api.handleCreateSession().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// TestHandleCreateSession_BadJSON невалидный JSON → 400
func TestHandleCreateSession_BadJSON(t *testing.T) {
	api := newTestAPI(newFakeUserRepo(), &fakePinger{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBufferString("{not json"))

	api.handleCreateSession().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleCreateSession_DBError ошибка БД → 500
func TestHandleCreateSession_DBError(t *testing.T) {
	repo := newFakeUserRepo()
	repo.findByEmailFn = func(_ context.Context, _ string) (*model.User, error) {
		return nil, errors.New("db error")
	}
	api := newTestAPI(repo, &fakePinger{})

	body := `{"email":"user@example.org","password":"secret123"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body))

	api.handleCreateSession().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

// --- GetUser ---
// TestHandleGetUser тесты found, not found, invalid id
func TestHandleGetUser(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		findByIDFn func(ctx context.Context, id int) (*model.User, error)
		wantCode   int
	}{
		{
			name:    "ok - user found",
			idParam: "42",
			findByIDFn: func(ctx context.Context, id int) (*model.User, error) {
				return &model.User{ID: id, Email: "user@example.org"}, nil
			},
			wantCode: http.StatusOK, // Найден → 200
		},
		{
			name:    "not found",
			idParam: "42",
			findByIDFn: func(_ context.Context, _ int) (*model.User, error) {
				return nil, nil
			},
			wantCode: http.StatusNotFound, // Не найден → 404
		},
		{
			name:       "invalid id",
			idParam:    "abc",
			findByIDFn: nil,                   // не вызывается
			wantCode:   http.StatusBadRequest, // Невалидный ID → 400
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepo()
			if tt.findByIDFn != nil {
				repo.findByIDFn = tt.findByIDFn
			}
			api := newTestAPI(repo, &fakePinger{})

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.idParam, nil)
			req.SetPathValue("id", tt.idParam)
			rec := httptest.NewRecorder()

			api.handleGetUser().ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("got %d, want %d, body=%q", rec.Code, tt.wantCode, rec.Body.String())
			}
		})
	}
}

// --- UpdateUser ---

// TestHandleUpdateUser_OK Успех → 200
func TestHandleUpdateUser_OK(t *testing.T) {
	repo := newFakeUserRepo()
	repo.updateFn = func(_ context.Context, u *model.User) (*model.User, error) {
		return u, nil
	}
	api := newTestAPI(repo, &fakePinger{})

	body := `{"email":"user@example.org","password":"secret123"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/users/42", strings.NewReader(body))
	req.SetPathValue("id", "42")

	api.handleUpdateUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestHandleUpdateUser_InvalidID Невалидный ID → 400
func TestHandleUpdateUser_InvalidID(t *testing.T) {
	api := newTestAPI(newFakeUserRepo(), &fakePinger{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/users/abc", nil)
	req.SetPathValue("id", "abc")

	api.handleUpdateUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleUpdateUser_ValidationError Валидация → 400
func TestHandleUpdateUser_ValidationError(t *testing.T) {
	repo := newFakeUserRepo()
	repo.updateFn = func(_ context.Context, _ *model.User) (*model.User, error) {
		return nil, model.ErrPasswordRequired
	}
	api := newTestAPI(repo, &fakePinger{})

	body := `{"email":"user@example.org","password":""}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/users/42", strings.NewReader(body))
	req.SetPathValue("id", "42")

	api.handleUpdateUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleUpdateUser_NotFound Пользователь не найден → 404
func TestHandleUpdateUser_NotFound(t *testing.T) {
	repo := newFakeUserRepo()
	// Настраиваем фейк так, чтобы он возвращал nil, nil
	repo.updateFn = func(_ context.Context, _ *model.User) (*model.User, error) {
		return nil, nil
	}
	api := newTestAPI(repo, &fakePinger{})

	body := `{"email":"user@example.org","password":"secret123"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/users/999", strings.NewReader(body))
	req.SetPathValue("id", "999")

	api.handleUpdateUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// --- DeleteUser ---

// TestHandleDeleteUser_OK Успех → 204
func TestHandleDeleteUser_OK(t *testing.T) {
	api := newTestAPI(newFakeUserRepo(), &fakePinger{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/42", nil)
	req.SetPathValue("id", "42")

	api.handleDeleteUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusNoContent)
	}
}

// TestHandleDeleteUser_InvalidID Невалидный ID → 400
func TestHandleDeleteUser_InvalidID(t *testing.T) {
	api := newTestAPI(newFakeUserRepo(), &fakePinger{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/abc", nil)
	req.SetPathValue("id", "abc")

	api.handleDeleteUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleDeleteUser_DBError Ошибка БД → 500
func TestHandleDeleteUser_DBError(t *testing.T) {
	repo := newFakeUserRepo()
	repo.deleteFn = func(_ context.Context, _ int) error {
		return errors.New("db error")
	}
	api := newTestAPI(repo, &fakePinger{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/42", nil)
	req.SetPathValue("id", "42")

	api.handleDeleteUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
