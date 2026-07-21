package auth_test

import (
	"testing"
	"time"

	"github.com/aeroxe/nexus-backend/pkg/auth"
)

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	jm := auth.NewJWTManager("test-secret", "nexus-test", 1*time.Hour, 7*24*time.Hour)

	pair, err := jm.GenerateTokenPair(42, 7, "user@test.com", []string{"admin"}, []string{"read", "write"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pair.AccessToken == "" {
		t.Errorf("expected non-empty access token")
	}
	if pair.RefreshToken == "" {
		t.Errorf("expected non-empty refresh token")
	}
	if pair.ExpiresIn != 3600 {
		t.Errorf("expected expires_in 3600, got %d", pair.ExpiresIn)
	}
}

func TestJWTManager_ValidateAccessToken_Success(t *testing.T) {
	jm := auth.NewJWTManager("test-secret", "nexus-test", 1*time.Hour, 7*24*time.Hour)

	pair, _ := jm.GenerateTokenPair(42, 7, "user@test.com", []string{"admin"}, []string{"read"})

	claims, err := jm.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("expected user ID 42, got %d", claims.UserID)
	}
	if claims.TenantID != 7 {
		t.Errorf("expected tenant ID 7, got %d", claims.TenantID)
	}
	if claims.Email != "user@test.com" {
		t.Errorf("expected email user@test.com, got %s", claims.Email)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Errorf("expected roles [admin], got %v", claims.Roles)
	}
}

func TestJWTManager_ValidateAccessToken_Invalid(t *testing.T) {
	jm := auth.NewJWTManager("test-secret", "nexus-test", 1*time.Hour, 7*24*time.Hour)

	_, err := jm.ValidateAccessToken("invalid.token.here")
	if err == nil {
		t.Fatalf("expected error for invalid token")
	}
}

func TestJWTManager_ValidateRefreshToken_Success(t *testing.T) {
	jm := auth.NewJWTManager("test-secret", "nexus-test", 1*time.Hour, 7*24*time.Hour)

	pair, _ := jm.GenerateTokenPair(42, 7, "user@test.com", []string{"admin"}, []string{"read"})

	claims, err := jm.ValidateRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if claims.Subject != "user@test.com" {
		t.Errorf("expected subject user@test.com, got %s", claims.Subject)
	}
}

func TestJWTManager_ValidateRefreshToken_Invalid(t *testing.T) {
	jm := auth.NewJWTManager("test-secret", "nexus-test", 1*time.Hour, 7*24*time.Hour)

	_, err := jm.ValidateRefreshToken("invalid.token.here")
	if err == nil {
		t.Fatalf("expected error for invalid refresh token")
	}
}
