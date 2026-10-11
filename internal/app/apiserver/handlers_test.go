package apiserver

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/model"
)

// --- фейки ---

type fakeUserRepo struct {
	createFn      func(ctx context.Context, u *model.User) (*model.User, error)
	findByIDFn    func(ctx context.Context, id int) (*model.User, error)
	findByEmailFn func(ctx context.Context, email string) (*model.User, error)
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

type fakePinger struct{ err error }

func (f *fakePinger) Ping(ctx context.Context) error { return f.err }

// --- хелпер ---

func newTestAPI(repo UserRepository, p Pinger) *APIServer {
	return New(NewConfig(), slog.Default(), repo, p)
}

// --- тесты ---

func TestHandleHealth_OK(t *testing.T) {
	api := newTestAPI(&fakeUserRepo{}, &fakePinger{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	api.handleHealth().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandleHealth_DBDown(t *testing.T) {
	api := newTestAPI(&fakeUserRepo{}, &fakePinger{err: context.DeadlineExceeded})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	api.handleHealth().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleGetUser_NotFound(t *testing.T) {
	api := newTestAPI(&fakeUserRepo{
		findByIDFn: func(_ context.Context, _ int) (*model.User, error) {
			return nil, nil
		},
	}, &fakePinger{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/42", nil)
	req.SetPathValue("id", "42")

	api.handleGetUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandleCreateUser_BadJSON(t *testing.T) {
	api := newTestAPI(&fakeUserRepo{}, &fakePinger{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{not json"))
	req.Header.Set("Content-Type", "application/json")

	api.handleCreateUser().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
