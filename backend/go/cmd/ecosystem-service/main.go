package main

import (
	"context"
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
	svcLogger := logger.New("ecosystem-service")
	svcLogger.Info("Starting Ecosystem Integration Service")

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

	integrationRepo := database.NewPostgresIntegrationRepository(db.Pool)
	toolRepo := database.NewPostgresMCPToolRepository(db.Pool)
	invocationRepo := database.NewPostgresMCPToolInvocationRepository(db.Pool)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy", "service": "ecosystem-service", "version": "1.0.0",
		})
	})
	mux.HandleFunc("/api/v1/integrations", listIntegrationsHandler(integrationRepo))
	mux.HandleFunc("/api/v1/integrations/create", createIntegrationHandler(integrationRepo))
	mux.HandleFunc("/api/v1/integrations/", integrationHandler(integrationRepo))
	mux.HandleFunc("/api/v1/mcp/tools", listMCPToolsHandler(toolRepo))
	mux.HandleFunc("/api/v1/mcp/tools/invoke", invokeMCPToolHandler(toolRepo, invocationRepo))

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8089")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Ecosystem Service listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func listIntegrationsHandler(repo *database.PostgresIntegrationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		ctx := r.Context()

		var integrations []*entities.Integration
		var err error

		if tenantID != "" {
			tid, e := strconv.ParseInt(tenantID, 10, 64)
			if e != nil {
				writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid tenant_id")
				return
			}
			integrations, err = repo.FindByTenantID(ctx, tid)
		} else {
			integrations = make([]*entities.Integration, 0)
		}

		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}
		if integrations == nil {
			integrations = make([]*entities.Integration, 0)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": integrations})
	}
}

func createIntegrationHandler(repo *database.PostgresIntegrationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
			return
		}

		var integ entities.Integration
		if err := json.NewDecoder(r.Body).Decode(&integ); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}
		integ.Status = "active"

		if err := repo.Create(r.Context(), &integ); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": integ})
	}
}

func integrationHandler(repo *database.PostgresIntegrationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/api/v1/integrations/"):]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid integration ID")
			return
		}

		integ, err := repo.FindByID(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Integration not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": integ})
	}
}

func listMCPToolsHandler(repo *database.PostgresMCPToolRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		ctx := r.Context()

		var tools []*entities.MCPTool
		var err error

		if tenantID != "" {
			tid, e := strconv.ParseInt(tenantID, 10, 64)
			if e != nil {
				writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid tenant_id")
				return
			}
			tools, err = repo.FindByTenantID(ctx, tid)
		} else {
			tools = make([]*entities.MCPTool, 0)
		}

		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}
		if tools == nil {
			tools = make([]*entities.MCPTool, 0)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": tools})
	}
}

func invokeMCPToolHandler(toolRepo *database.PostgresMCPToolRepository, invocationRepo *database.PostgresMCPToolInvocationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
			return
		}

		var call struct {
			ToolID    int64                  `json:"tool_id"`
			TenantID  int64                  `json:"tenant_id"`
			UserID    int64                  `json:"user_id"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		ctx := r.Context()

		tool, err := toolRepo.FindByID(ctx, call.ToolID)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP tool not found")
			return
		}

		if tool.Status != "active" {
			writeError(w, http.StatusForbidden, "TOOL_DISABLED", "This MCP tool is disabled")
			return
		}

		inputJSON, _ := json.Marshal(call.Arguments)
		start := time.Now()

		invocation := &entities.MCPToolInvocation{
			ToolID:   tool.ID,
			TenantID: call.TenantID,
			UserID:   call.UserID,
			Input:    string(inputJSON),
			Status:   "success",
		}

		_ = simulateToolExecution(tool, call.Arguments)

		invocation.LatencyMs = float64(time.Since(start).Milliseconds())
		invocation.Output = fmt.Sprintf("Tool '%s' invoked with arguments", tool.Name)

		invocationRepo.Create(context.Background(), invocation)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"tool_id": tool.ID,
				"tool":    tool.Name,
				"status":  "invoked",
				"result":  invocation.Output,
			},
		})
	}
}

func simulateToolExecution(tool *entities.MCPTool, args map[string]interface{}) error {
	return nil
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
