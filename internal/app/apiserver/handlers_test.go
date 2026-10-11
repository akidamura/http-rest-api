package apiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/store"
)

func TestAPIServer_HandleHealth(t *testing.T) {
	// 1. Поднимаем Store вручную (NewTestStore недоступен)
	cfg := store.NewConfig()
	st := store.New(cfg)
	if err := st.Open(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })

	// 2. Собираем APIServer с этим Store
	api := New(NewConfig())
	api.store = st

	// 3. Дёргаем handler
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	api.handleHealth().ServeHTTP(rec, req)

	// 4. Проверяем ответ
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
