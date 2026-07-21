package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

type Integration struct {
	ID          string                 `json:"id"`
	TenantID    int64                  `json:"tenant_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Category    string                 `json:"category"`
	Status      string                 `json:"status"`
	Config      map[string]interface{} `json:"config"`
	Scopes      []string               `json:"scopes,omitempty"`
	WebhookURL  string                 `json:"webhook_url,omitempty"`
	Description string                 `json:"description"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

type MCPTool struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"input_schema"`
	ServerID    string                 `json:"server_id"`
	Enabled     bool                   `json:"enabled"`
}

type MCPToolCall struct {
	ToolID    string                 `json:"tool_id"`
	Arguments map[string]interface{} `json:"arguments"`
}

type EcosystemStore struct {
	integrations map[string]*Integration
	mcpTools     map[string]*MCPTool
	mu           sync.RWMutex
}

var store = &EcosystemStore{
	integrations: make(map[string]*Integration),
	mcpTools:     make(map[string]*MCPTool),
}

func init() {
	store.integrations["int-001"] = &Integration{ID: "int-001", TenantID: 1, Name: "GitHub", Type: "git", Category: "development", Status: "active", Config: map[string]interface{}{"repo_url": "https://github.com/example"}, Description: "GitHub integration for code repos", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)}
	store.integrations["int-002"] = &Integration{ID: "int-002", TenantID: 1, Name: "Slack", Type: "messaging", Category: "communication", Status: "active", Config: map[string]interface{}{"workspace": "aeroxe"}, Description: "Slack notifications", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)}
	store.integrations["int-003"] = &Integration{ID: "int-003", TenantID: 1, Name: "PostgreSQL", Type: "database", Category: "data", Status: "active", Config: map[string]interface{}{"host": "localhost", "port": 5432}, Description: "Primary database", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)}

	store.mcpTools["mcp-001"] = &MCPTool{ID: "mcp-001", Name: "execute_sql", Description: "Execute SQL queries on connected databases", Schema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}}, ServerID: "mcp-server-db", Enabled: true}
	store.mcpTools["mcp-002"] = &MCPTool{ID: "mcp-002", Name: "search_code", Description: "Search code repositories", Schema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}}, ServerID: "mcp-server-git", Enabled: true}
	store.mcpTools["mcp-003"] = &MCPTool{ID: "mcp-003", Name: "send_message", Description: "Send messages to channels", Schema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"channel": map[string]interface{}{"type": "string"}, "message": map[string]interface{}{"type": "string"}}}, ServerID: "mcp-server-chat", Enabled: true}
}

func main() {
	svcLogger := logger.New("ecosystem-service")
	svcLogger.Info("Starting Ecosystem Integration Service")

	cfg, err := config.LoadConfig("")
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to load config: %v", err))
	}
	_ = cfg

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/integrations", listIntegrationsHandler)
	mux.HandleFunc("/api/v1/integrations/create", createIntegrationHandler)
	mux.HandleFunc("/api/v1/integrations/", integrationHandler)
	mux.HandleFunc("/api/v1/mcp/tools", listMCPToolsHandler)
	mux.HandleFunc("/api/v1/mcp/tools/invoke", invokeMCPToolHandler)

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8089")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Ecosystem Service listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "healthy", "service": "ecosystem-service", "version": "1.0.0",
	})
}

func listIntegrationsHandler(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	integrations := make([]*Integration, 0)
	for _, i := range store.integrations {
		integrations = append(integrations, i)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": integrations})
}

func createIntegrationHandler(w http.ResponseWriter, r *http.Request) {
	var integ Integration
	if err := json.NewDecoder(r.Body).Decode(&integ); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	integ.ID = fmt.Sprintf("int-%d", time.Now().UnixNano())
	integ.Status = "active"
	integ.CreatedAt = time.Now().Format(time.RFC3339)
	integ.UpdatedAt = time.Now().Format(time.RFC3339)

	store.mu.Lock()
	store.integrations[integ.ID] = &integ
	store.mu.Unlock()

	writeJSON(w, http.StatusCreated, map[string]interface{}{"data": integ})
}

func integrationHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/integrations/"):]

	store.mu.RLock()
	integ, exists := store.integrations[id]
	store.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Integration not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": integ})
}

func listMCPToolsHandler(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	tools := make([]*MCPTool, 0)
	for _, t := range store.mcpTools {
		tools = append(tools, t)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": tools})
}

func invokeMCPToolHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var call MCPToolCall
	if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	store.mu.RLock()
	tool, exists := store.mcpTools[call.ToolID]
	store.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP tool not found")
		return
	}

	if !tool.Enabled {
		writeError(w, http.StatusForbidden, "TOOL_DISABLED", "This MCP tool is disabled")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"tool_id": call.ToolID,
			"tool":    tool.Name,
			"status":  "invoked",
			"result":  fmt.Sprintf("Tool '%s' invoked with arguments", tool.Name),
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
