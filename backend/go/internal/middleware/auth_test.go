package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/auth"
)

func newTestJWTManager() *auth.JWTManager {
	return auth.NewJWTManager("test-secret-key", "nexus-test", 1*time.Hour, 7*24*time.Hour)
}

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}

func TestAuthMiddleware_NoHeader(t *testing.T) {
	jwtManager := newTestJWTManager()
	mw := middleware.AuthMiddleware(jwtManager)
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	jwtManager := newTestJWTManager()
	mw := middleware.AuthMiddleware(jwtManager)
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtManager := newTestJWTManager()
	mw := middleware.AuthMiddleware(jwtManager)
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	jwtManager := newTestJWTManager()
	tokenPair, err := jwtManager.GenerateTokenPair(42, 7, "user@test.com", []string{"admin"}, []string{"read", "write"})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	mw := middleware.AuthMiddleware(jwtManager)

	var capturedReq *http.Request
	captureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(captureHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedReq == nil {
		t.Fatalf("expected request to be captured")
	}
}

func TestAuthMiddleware_GetUserID(t *testing.T) {
	jwtManager := newTestJWTManager()
	tokenPair, _ := jwtManager.GenerateTokenPair(42, 7, "user@test.com", []string{"admin"}, []string{"read"})

	mw := middleware.AuthMiddleware(jwtManager)
	var capturedReq *http.Request
	captureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	})
	handler := mw(captureHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	userID := middleware.GetUserID(capturedReq)
	if userID != 42 {
		t.Errorf("expected user ID 42, got %d", userID)
	}
}

func TestAuthMiddleware_GetTenantID(t *testing.T) {
	jwtManager := newTestJWTManager()
	tokenPair, _ := jwtManager.GenerateTokenPair(42, 7, "user@test.com", []string{"admin"}, []string{"read"})

	mw := middleware.AuthMiddleware(jwtManager)
	var capturedReq *http.Request
	captureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	})
	handler := mw(captureHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	tenantID := middleware.GetTenantID(capturedReq)
	if tenantID != 7 {
		t.Errorf("expected tenant ID 7, got %d", tenantID)
	}
}

func TestRequirePermission_Allowed(t *testing.T) {
	jwtManager := newTestJWTManager()
	tokenPair, _ := jwtManager.GenerateTokenPair(1, 1, "u@test.com", []string{"user"}, []string{"read", "write"})

	authMw := middleware.AuthMiddleware(jwtManager)
	permMw := middleware.RequirePermission("read")
	handler := authMw(permMw(newTestHandler()))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRequirePermission_Denied(t *testing.T) {
	jwtManager := newTestJWTManager()
	tokenPair, _ := jwtManager.GenerateTokenPair(1, 1, "u@test.com", []string{"user"}, []string{"read"})

	authMw := middleware.AuthMiddleware(jwtManager)
	permMw := middleware.RequirePermission("delete")
	handler := authMw(permMw(newTestHandler()))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}
