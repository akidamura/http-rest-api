package model_test

import (
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/model"
)

func TestUser_BeforeCreate(t *testing.T) {
	u := model.NewTestUser(t)

	if err := u.BeforeCreate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.EncryptedPassword == "" {
		t.Error("expected encrypted password to be set")
	}
	if u.Password != "" {
		t.Error("expected password to be cleared")
	}
}
