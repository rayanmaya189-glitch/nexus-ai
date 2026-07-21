package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/infrastructure/database"
	"github.com/aeroxe/nexus-backend/internal/interfaces/rest"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

func main() {
	svcLogger := logger.New("audit-service")
	svcLogger.Info("Starting Audit Service")

	cfg, err := config.LoadConfig("")
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to load config: %v", err))
	}

	db, err := database.NewDB(
		getEnv("DB_HOST", cfg.Database.Host),
		getEnvInt("DB_PORT", cfg.Database.Port),
		getEnv("DB_USER", cfg.Database.User),
		getEnv("DB_PASSWORD", cfg.Database.Password),
		getEnv("DB_NAME", cfg.Database.DBName),
		getEnv("DB_SSLMODE", cfg.Database.SSLMode),
	)
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer db.Close()
	svcLogger.Info("Connected to PostgreSQL database")

	auditRepo := database.NewPostgresAuditLogRepository(db.Pool)

	auditUseCase := usecases.NewAuditUseCase(auditRepo)
	auditHandler := rest.NewAuditHandler(auditUseCase)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"audit-service","version":"1.0.0"}`)
	})

	mux.HandleFunc("/api/v1/audit/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			auditHandler.ListAuditLogs(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/v1/audit/logs/create", auditHandler.CreateAuditLog)
	mux.HandleFunc("/api/v1/audit/logs/", auditHandler.GetAuditLog)

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8085")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Audit Service listening on %s", addr))

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
