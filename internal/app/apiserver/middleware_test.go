package apiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/model"
)

// makeAuthCookie создаёт валидную куку с указанным user_id.
func makeAuthCookie(t *testing.T, api *APIServer, userID int) *http.Cookie {
	t.Helper()

	// Создаём фейковый запрос, чтобы получить сессию
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	session, err := api.sessionStore.Get(req, sessionName)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}

	session.Values["user_id"] = userID

	// Сохраняем сессию в фейковый ResponseWriter
	rec := httptest.NewRecorder()
	if err := session.Save(req, rec); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	// Достаём куку из Set-Cookie
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cookie to be set")
	}
	return cookies[0]
}

// --- Тесты authenticateUser ---

func TestAuthenticateUser(t *testing.T) {
	user := &model.User{ID: 42, Email: "user@example.org"}

	tests := []struct {
		name         string
		cookie       func(t *testing.T, api *APIServer) *http.Cookie
		findByIDFn   func(ctx context.Context, id int) (*model.User, error)
		wantCode     int
		wantNextCall bool
	}{
		{
			name: "ok — valid cookie",
			cookie: func(t *testing.T, api *APIServer) *http.Cookie {
				return makeAuthCookie(t, api, user.ID)
			},
			findByIDFn: func(_ context.Context, id int) (*model.User, error) {
				return user, nil
			},
			wantCode:     http.StatusOK,
			wantNextCall: true,
		},
		{
			name:   "no cookie",
			cookie: func(t *testing.T, api *APIServer) *http.Cookie { return nil },
			findByIDFn: func(_ context.Context, _ int) (*model.User, error) {
				return user, nil
			},
			wantCode:     http.StatusUnauthorized,
			wantNextCall: false,
		},
		{
			name: "invalid cookie",
			cookie: func(t *testing.T, api *APIServer) *http.Cookie {
				return &http.Cookie{
					Name:  sessionName,
					Value: "fake-invalid-cookie",
				}
			},
			findByIDFn: func(_ context.Context, _ int) (*model.User, error) {
				return user, nil
			},
			wantCode:     http.StatusUnauthorized,
			wantNextCall: false,
		},
		{
			name: "user not found in DB",
			cookie: func(t *testing.T, api *APIServer) *http.Cookie {
				return makeAuthCookie(t, api, 999)
			},
			findByIDFn: func(_ context.Context, _ int) (*model.User, error) {
				return nil, nil // не найден
			},
			wantCode:     http.StatusUnauthorized,
			wantNextCall: false,
		},
		{
			name: "DB error",
			cookie: func(t *testing.T, api *APIServer) *http.Cookie {
				return makeAuthCookie(t, api, user.ID)
			},
			findByIDFn: func(_ context.Context, _ int) (*model.User, error) {
				return nil, errors.New("db error")
			},
			wantCode:     http.StatusInternalServerError,
			wantNextCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepo()
			repo.findByIDFn = tt.findByIDFn
			api := newTestAPI(repo, &fakePinger{})

			// next фиксирует вызов + user в контексте
			var nextCalled bool
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				got, ok := r.Context().Value(ctxKeyUser).(*model.User)
				if !ok {
					t.Error("expected user in context")
					return
				}
				if got.ID != user.ID {
					t.Errorf("expected user ID %d, got %d", user.ID, got.ID)
				}
			})

			// Собираем запрос с кукой
			req := httptest.NewRequest(http.MethodGet, "/private/whoami", nil)
			if c := tt.cookie(t, api); c != nil {
				req.AddCookie(c)
			}

			rec := httptest.NewRecorder()
			api.authenticateUser(next).ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("got code %d, want %d, body=%q", rec.Code, tt.wantCode, rec.Body.String())
			}
			if nextCalled != tt.wantNextCall {
				t.Errorf("got nextCalled=%v, want %v", nextCalled, tt.wantNextCall)
			}
		})
	}
}

// --- Тесты handleWhoami ---

func TestHandleWhoami(t *testing.T) {
	user := &model.User{ID: 42, Email: "user@example.org"}

	tests := []struct {
		name     string
		ctx      context.Context
		wantCode int
	}{
		{
			name:     "ok — user in context",
			ctx:      context.WithValue(context.Background(), ctxKeyUser, user),
			wantCode: http.StatusOK,
		},
		{
			name:     "not authenticated — no user in context",
			ctx:      context.Background(),
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newTestAPI(newFakeUserRepo(), &fakePinger{})

			req := httptest.NewRequest(http.MethodGet, "/private/whoami", nil).WithContext(tt.ctx)
			rec := httptest.NewRecorder()

			api.handleWhoami().ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("got %d, want %d, body=%q", rec.Code, tt.wantCode, rec.Body.String())
			}
		})
	}
}
