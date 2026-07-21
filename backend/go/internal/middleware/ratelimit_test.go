package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/aeroxe/nexus-backend/internal/middleware"
)

func TestRateLimitMiddleware_WithinLimit(t *testing.T) {
	rl := middleware.NewRateLimitMiddleware(5, 1*time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	limit := rec.Header().Get("X-RateLimit-Limit")
	if limit != "5" {
		t.Errorf("expected X-RateLimit-Limit 5, got %s", limit)
	}

	remaining := rec.Header().Get("X-RateLimit-Remaining")
	if remaining != "4" {
		t.Errorf("expected X-RateLimit-Remaining 4, got %s", remaining)
	}
}

func TestRateLimitMiddleware_ExceedsLimit(t *testing.T) {
	rl := middleware.NewRateLimitMiddleware(3, 1*time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:5678"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:5678"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_Headers(t *testing.T) {
	rl := middleware.NewRateLimitMiddleware(10, 1*time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "172.16.0.1:9999"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	limit := rec.Header().Get("X-RateLimit-Limit")
	parsed, err := strconv.Atoi(limit)
	if err != nil || parsed != 10 {
		t.Errorf("expected X-RateLimit-Limit 10, got %s", limit)
	}

	remaining := rec.Header().Get("X-RateLimit-Remaining")
	parsedRem, err := strconv.Atoi(remaining)
	if err != nil || parsedRem != 9 {
		t.Errorf("expected X-RateLimit-Remaining 9, got %s", remaining)
	}

	reset := rec.Header().Get("X-RateLimit-Reset")
	if reset == "" {
		t.Errorf("expected X-RateLimit-Reset to be set")
	}
}
