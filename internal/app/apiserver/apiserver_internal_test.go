package apiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIServer_HandleHello(t *testing.T) {
	s := New(NewConfig())
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/hello", nil)

	s.handleHello().ServeHTTP(rec, req)

	if rec.Body.String() != "Hello" {
		t.Errorf("expect 'Hello', but got '%s'", rec.Body.String())
	}

}
