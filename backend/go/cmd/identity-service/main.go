package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/infrastructure/database"
	"github.com/aeroxe/nexus-backend/internal/interfaces/rest"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/auth"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

func main() {
	svcLogger := logger.New("identity-service")
	svcLogger.Info("Starting Identity Service")

	cfg, err := config.LoadConfig("")
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to load config: %v", err))
	}

	dbHost := getEnv("DB_HOST", cfg.Database.Host)
	dbPort := getEnvInt("DB_PORT", cfg.Database.Port)
	dbUser := getEnv("DB_USER", cfg.Database.User)
	dbPass := getEnv("DB_PASSWORD", cfg.Database.Password)
	dbName := getEnv("DB_NAME", cfg.Database.DBName)
	dbSSL := getEnv("DB_SSLMODE", cfg.Database.SSLMode)

	db, err := database.NewDB(dbHost, dbPort, dbUser, dbPass, dbName, dbSSL)
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer db.Close()
	svcLogger.Info("Connected to PostgreSQL database")

	userRepo := database.NewPostgresUserRepository(db.Pool)
	tenantRepo := database.NewPostgresTenantRepository(db.Pool)
	roleRepo := database.NewPostgresRoleRepository(db.Pool)

	jwtManager := auth.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	authUseCase := usecases.NewAuthUseCase(userRepo, tenantRepo, roleRepo, jwtManager)
	authHandler := rest.NewAuthHandler(authUseCase)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"identity-service","version":"1.0.0"}`)
	})

	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.RefreshToken)

	mux.HandleFunc("/api/v1/users", authHandler.ListUsers)
	mux.HandleFunc("/api/v1/users/create", authHandler.CreateUser)
	mux.HandleFunc("/api/v1/users/update", authHandler.UpdateUser)
	mux.HandleFunc("/api/v1/users/", authHandler.GetUser)

	mux.HandleFunc("/api/v1/tenants", authHandler.ListTenants)
	mux.HandleFunc("/api/v1/tenants/create", authHandler.CreateTenant)
	mux.HandleFunc("/api/v1/tenants/update", authHandler.UpdateTenant)
	mux.HandleFunc("/api/v1/tenants/", authHandler.DeleteTenant)

	mux.HandleFunc("/api/v1/roles", authHandler.ListRoles)
	mux.HandleFunc("/api/v1/roles/create", authHandler.CreateRole)
	mux.HandleFunc("/api/v1/roles/", authHandler.DeleteRole)

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8081")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Identity Service listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var n int
		fmt.Sscanf(val, "%d", &n)
		if n > 0 {
			return n
		}
	}
	return defaultVal
}
