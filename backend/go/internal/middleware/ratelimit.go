package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimitMiddleware struct {
	visits  map[string]*visitRecord
	mu      sync.Mutex
	limit   int
	window  time.Duration
}

type visitRecord struct {
	count    int
	firstAt  time.Time
}

func NewRateLimitMiddleware(limit int, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		visits: make(map[string]*visitRecord),
		limit:  limit,
		window: window,
	}
}

func (m *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		now := time.Now()

		m.mu.Lock()
		rec, exists := m.visits[key]
		if !exists || now.Sub(rec.firstAt) > m.window {
			m.visits[key] = &visitRecord{count: 1, firstAt: now}
			m.mu.Unlock()

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(m.limit-1))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(int64(m.window.Seconds()), 10))
			next.ServeHTTP(w, r)
			return
		}

		rec.count++
		remaining := m.limit - rec.count
		m.mu.Unlock()

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, remaining)))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(int64(m.window.Seconds()), 10))

		if remaining < 0 {
			retryAfter := int(m.window.Seconds())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": fmt.Sprintf("Rate limit exceeded. Retry after %d seconds.", retryAfter),
				},
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
