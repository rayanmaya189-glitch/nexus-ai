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
	svcLogger := logger.New("workflow-service")
	svcLogger.Info("Starting Workflow Service")

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

	wfRepo := database.NewPostgresWorkflowRepository(db.Pool)
	stepRepo := database.NewPostgresWorkflowStepRepository(db.Pool)

	wfUseCase := usecases.NewWorkflowUseCase(wfRepo, stepRepo)
	wfHandler := rest.NewWorkflowHandler(wfUseCase)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"workflow-service","version":"1.0.0"}`)
	})

	mux.HandleFunc("/api/v1/workflows", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			wfHandler.ListWorkflows(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/v1/workflows/create", wfHandler.CreateWorkflow)
	mux.HandleFunc("/api/v1/workflows/update", wfHandler.UpdateWorkflow)
	mux.HandleFunc("/api/v1/workflows/delete", wfHandler.DeleteWorkflow)
	mux.HandleFunc("/api/v1/workflows/steps", wfHandler.GetWorkflowSteps)
	mux.HandleFunc("/api/v1/workflows/", wfHandler.GetWorkflow)

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8084")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Workflow Service listening on %s", addr))

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
