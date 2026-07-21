package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/infrastructure/database"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

func main() {
	svcLogger := logger.New("model-registry")
	svcLogger.Info("Starting Model Registry Service")

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

	modelRepo := database.NewPostgresAIModelRepository(db.Pool)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/models", listModelsHandler(modelRepo))
	mux.HandleFunc("/api/v1/models/create", createModelHandler(modelRepo))
	mux.HandleFunc("/api/v1/models/metrics/", metricsHandler)
	mux.HandleFunc("/api/v1/models/pull/", pullModelHandler)
	mux.HandleFunc("/api/v1/models/", modelHandler(modelRepo))

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8086")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Model Registry listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "healthy", "service": "model-registry", "version": "1.0.0",
	})
}

func listModelsHandler(repo *database.PostgresAIModelRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		category := r.URL.Query().Get("category")

		var models []*entities.AIModel
		var err error

		if category != "" {
			models, err = repo.FindByCategory(r.Context(), category)
		} else {
			models, err = repo.List(r.Context())
		}

		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}
		if models == nil {
			models = make([]*entities.AIModel, 0)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": models})
	}
}

func createModelHandler(repo *database.PostgresAIModelRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
			return
		}

		var model entities.AIModel
		if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		model.Status = "active"
		if model.CreatedAt.IsZero() {
			model.CreatedAt = time.Now()
		}

		if err := repo.Create(r.Context(), &model); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": model})
	}
}

func modelHandler(repo *database.PostgresAIModelRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/api/v1/models/"):]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid model ID")
			return
		}

		model, err := repo.FindByID(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Model not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": model})
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	modelID := r.URL.Path[len("/api/v1/models/metrics/"):]
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"model_id":       modelID,
			"request_count":  0,
			"avg_latency_ms": 0,
			"error_rate":     0,
		},
	})
}

func pullModelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	modelID := r.URL.Path[len("/api/v1/models/pull/"):]

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"model_id": modelID,
			"status":   "pulling",
			"message":  fmt.Sprintf("Pulling model %s from Ollama", modelID),
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
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
