package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/aeroxe/nexus-backend/pkg/auth"
	"github.com/aeroxe/nexus-backend/pkg/httputil"
)

type contextKey string

const (
	UserIDKey      contextKey = "user_id"
	TenantIDKey    contextKey = "tenant_id"
	EmailKey       contextKey = "email"
	RolesKey       contextKey = "roles"
	PermissionsKey contextKey = "permissions"
)

func AuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization format")
				return
			}

			claims, err := jwtManager.ValidateAccessToken(parts[1])
			if err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, RolesKey, claims.Roles)
			ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) int64 {
	if v, ok := r.Context().Value(UserIDKey).(int64); ok {
		return v
	}
	return 0
}

func GetTenantID(r *http.Request) int64 {
	if v, ok := r.Context().Value(TenantIDKey).(int64); ok {
		return v
	}
	return 0
}

func GetEmail(r *http.Request) string {
	if v, ok := r.Context().Value(EmailKey).(string); ok {
		return v
	}
	return ""
}

func GetRoles(r *http.Request) []string {
	if v, ok := r.Context().Value(RolesKey).([]string); ok {
		return v
	}
	return nil
}

func GetPermissions(r *http.Request) []string {
	if v, ok := r.Context().Value(PermissionsKey).([]string); ok {
		return v
	}
	return nil
}

func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions := GetPermissions(r)
			for _, p := range permissions {
				if p == permission || p == "*" {
					next.ServeHTTP(w, r)
					return
				}
			}
			httputil.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
		})
	}
}
