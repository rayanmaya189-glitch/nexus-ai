package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aeroxe/nexus-backend/internal/middleware"
)

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	handler := middleware.RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Errorf("expected X-Request-ID header to be set")
	}
	if len(requestID) < 10 {
		t.Errorf("expected X-Request-ID to be a UUID-like string, got %s", requestID)
	}
}

func TestRequestIDMiddleware_PreservesExistingID(t *testing.T) {
	handler := middleware.RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "my-custom-request-id")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID != "my-custom-request-id" {
		t.Errorf("expected X-Request-ID my-custom-request-id, got %s", requestID)
	}
}
