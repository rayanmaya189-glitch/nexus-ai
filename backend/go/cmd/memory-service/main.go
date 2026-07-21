package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/infrastructure/database"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

func main() {
	svcLogger := logger.New("memory-service")
	svcLogger.Info("Starting Memory Service")

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

	memoryRepo := database.NewPostgresMemoryRepository(db.Pool)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy", "service": "memory-service", "version": "1.0.0",
		})
	})

	mux.HandleFunc("/api/v1/memory", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listMemoryHandler(memoryRepo)(w, r)
		case http.MethodPost:
			createMemoryHandler(memoryRepo)(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only")
		}
	})

	mux.HandleFunc("/api/v1/memory/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			updateMemoryHandler(memoryRepo)(w, r)
		case http.MethodDelete:
			deleteMemoryHandler(memoryRepo)(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "PUT or DELETE only")
		}
	})

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8091")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Memory Service listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func listMemoryHandler(repo *database.PostgresMemoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		agentID := r.URL.Query().Get("agent_id")
		userID := r.URL.Query().Get("user_id")
		tenantID := r.URL.Query().Get("tenant_id")

		switch {
		case agentID != "":
			aid, err := strconv.ParseInt(agentID, 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid agent_id")
				return
			}
			memories, err := repo.FindByAgentID(ctx, aid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
				return
			}
			if memories == nil {
				memories = make([]*entities.Memory, 0)
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": memories})
		case userID != "" && tenantID != "":
			tid, _ := strconv.ParseInt(tenantID, 10, 64)
			uid, _ := strconv.ParseInt(userID, 10, 64)
			memories, err := repo.FindByTenantAndUser(ctx, tid, uid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
				return
			}
			if memories == nil {
				memories = make([]*entities.Memory, 0)
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": memories})
		default:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": []*entities.Memory{}})
		}
	}
}

func createMemoryHandler(repo *database.PostgresMemoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mem entities.Memory
		if err := json.NewDecoder(r.Body).Decode(&mem); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		if err := repo.Create(r.Context(), &mem); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": mem})
	}
}

func updateMemoryHandler(repo *database.PostgresMemoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/api/v1/memory/"):]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid memory ID")
			return
		}

		existing, err := repo.FindByID(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Memory not found")
			return
		}

		var updated entities.Memory
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		if updated.Content != "" {
			existing.Content = updated.Content
		}
		if updated.Summary != "" {
			existing.Summary = updated.Summary
		}
		if updated.Importance > 0 {
			existing.Importance = updated.Importance
		}
		existing.AccessCount = updated.AccessCount

		if err := repo.Update(r.Context(), existing); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": existing})
	}
}

func deleteMemoryHandler(repo *database.PostgresMemoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/api/v1/memory/"):]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid memory ID")
			return
		}

		if err := repo.Delete(r.Context(), id); err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Memory not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": map[string]bool{"deleted": true}})
	}
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
