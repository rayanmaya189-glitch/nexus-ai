package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyTenantID contextKey = "tenant_id"
	ContextKeyEmail    contextKey = "email"
	ContextKeyRoles    contextKey = "roles"
)

type Claims struct {
	UserID      int64    `json:"user_id"`
	TenantID    int64    `json:"tenant_id"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// Public paths that don't require authentication
var publicPaths = map[string]bool{
	"/health":                   true,
	"/api/v1/auth/login":        true,
	"/api/v1/auth/register":     true,
	"/api/v1/auth/refresh":      true,
	"/":                         true,
}

func authMiddleware(jwtSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public paths
		if publicPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for health checks
		if strings.HasPrefix(r.URL.Path, "/health") {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Allow API key authentication
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" {
				r = r.WithContext(context.WithValue(r.Context(), ContextKeyUserID, int64(0)))
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error":{"code":"UNAUTHORIZED","message":"Missing authorization header"}}`)
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid authorization format"}}`)
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid or expired token"}}`)
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid token claims"}}`)
			return
		}

		// Check expiration
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error":{"code":"UNAUTHORIZED","message":"Token expired"}}`)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeyTenantID, claims.TenantID)
		ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
		ctx = context.WithValue(ctx, ContextKeyRoles, claims.Roles)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
