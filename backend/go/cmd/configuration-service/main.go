package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type PlatformConfig struct {
	ID          string                 `json:"id"`
	TenantID    int64                  `json:"tenant_id"`
	Category    string                 `json:"category"`
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	ValueType   string                 `json:"value_type"`
	Encrypted   bool                   `json:"encrypted"`
	Editable    bool                   `json:"editable"`
	Description string                 `json:"description"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

type ConfigStore struct {
	configs map[string]*PlatformConfig
	mu      sync.RWMutex
}

var store = &ConfigStore{configs: make(map[string]*PlatformConfig)}

func init() {
	defaults := []*PlatformConfig{
		{ID: "cfg-001", TenantID: 0, Category: "ai", Key: "default_model", Value: "phi4-mini:3.8b", ValueType: "string", Encrypted: false, Editable: true, Description: "Default AI model for chat", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
		{ID: "cfg-002", TenantID: 0, Category: "ai", Key: "max_tokens", Value: 4096, ValueType: "int", Encrypted: false, Editable: true, Description: "Maximum tokens per request", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
		{ID: "cfg-003", TenantID: 0, Category: "ai", Key: "temperature", Value: 0.7, ValueType: "float", Encrypted: false, Editable: true, Description: "Default temperature", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
		{ID: "cfg-004", TenantID: 0, Category: "security", Key: "sensitive_filter_enabled", Value: true, ValueType: "bool", Encrypted: false, Editable: true, Description: "Enable sensitive words filter", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
		{ID: "cfg-005", TenantID: 0, Category: "security", Key: "injection_detection_enabled", Value: true, ValueType: "bool", Encrypted: false, Editable: true, Description: "Enable prompt injection detection", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
		{ID: "cfg-006", TenantID: 0, Category: "audit", Key: "retention_days", Value: 730, ValueType: "int", Encrypted: false, Editable: true, Description: "Audit log retention in days (2 years)", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
		{ID: "cfg-007", TenantID: 0, Category: "rag", Key: "chunk_size", Value: 512, ValueType: "int", Encrypted: false, Editable: true, Description: "Default RAG chunk size", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
		{ID: "cfg-008", TenantID: 0, Category: "rag", Key: "chunk_overlap", Value: 50, ValueType: "int", Encrypted: false, Editable: true, Description: "Default RAG chunk overlap", CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
	}

	store.mu.Lock()
	for _, c := range defaults {
		store.configs[c.ID] = c
	}
	store.mu.Unlock()
}

func main() {
	log.Println("Starting Configuration Service")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/config", listConfigHandler)
	mux.HandleFunc("/api/v1/config/create", createConfigHandler)
	mux.HandleFunc("/api/v1/config/", configHandler)

	port := getEnv("PORT", "8088")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Configuration Service listening on %s", addr)

	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "healthy", "service": "configuration-service", "version": "1.0.0",
	})
}

func listConfigHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	category := r.URL.Query().Get("category")

	store.mu.RLock()
	defer store.mu.RUnlock()

	configs := make([]*PlatformConfig, 0)
	for _, c := range store.configs {
		if tenantID != "" && fmt.Sprintf("%d", c.TenantID) != tenantID {
			continue
		}
		if category != "" && c.Category != category {
			continue
		}
		configs = append(configs, c)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": configs})
}

func createConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var cfg PlatformConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	cfg.ID = fmt.Sprintf("cfg-%d", time.Now().UnixNano())
	cfg.CreatedAt = time.Now().Format(time.RFC3339)
	cfg.UpdatedAt = time.Now().Format(time.RFC3339)

	store.mu.Lock()
	store.configs[cfg.ID] = &cfg
	store.mu.Unlock()

	writeJSON(w, http.StatusCreated, map[string]interface{}{"data": cfg})
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/config/"):]

	store.mu.RLock()
	cfg, exists := store.configs[id]
	store.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Configuration not found")
		return
	}

	if r.Method == http.MethodPut {
		if !cfg.Editable {
			writeError(w, http.StatusForbidden, "NOT_EDITABLE", "This configuration is not editable")
			return
		}

		var updated PlatformConfig
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		cfg.Value = updated.Value
		cfg.UpdatedAt = time.Now().Format(time.RFC3339)

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": cfg})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": cfg})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		next.ServeHTTP(w, r)
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
	if v := os.Getenv(key); v != "" { return v }
	return def
}
