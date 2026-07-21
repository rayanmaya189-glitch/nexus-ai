package auth_test

import (
	"testing"

	"github.com/aeroxe/nexus-backend/pkg/auth"
)

func TestHashPassword(t *testing.T) {
	hash, err := auth.HashPassword("mypassword")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash == "" {
		t.Errorf("expected non-empty hash")
	}
	if hash == "mypassword" {
		t.Errorf("hash should not equal plaintext password")
	}
}

func TestCheckPassword_Correct(t *testing.T) {
	password := "securepassword123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !auth.CheckPassword(password, hash) {
		t.Errorf("expected password check to succeed")
	}
}

func TestCheckPassword_Incorrect(t *testing.T) {
	hash, err := auth.HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if auth.CheckPassword("wrongpassword", hash) {
		t.Errorf("expected password check to fail")
	}
}
