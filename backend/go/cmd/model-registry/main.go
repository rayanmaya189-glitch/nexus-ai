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

type AIModel struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ModelID        string   `json:"model_id"`
	Category       string   `json:"category"`
	SizeBytes      int64    `json:"size_bytes"`
	Parameters     string   `json:"parameters"`
	ContextLength  int      `json:"context_length"`
	Capabilities   []string `json:"capabilities"`
	Status         string   `json:"status"`
	TenantID       *int64   `json:"tenant_id,omitempty"`
	MaxConcurrency int      `json:"max_concurrency"`
	RequestCount   int64    `json:"request_count"`
	AvgLatencyMs   float64  `json:"avg_latency_ms"`
	CreatedAt      string   `json:"created_at"`
}

type ModelMetrics struct {
	ModelID       string  `json:"model_id"`
	RequestCount  int64   `json:"request_count"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	ErrorRate     float64 `json:"error_rate"`
	ActiveWorkers int     `json:"active_workers"`
}

type ModelRegistryStore struct {
	models  map[string]*AIModel
	metrics map[string]*ModelMetrics
	mu      sync.RWMutex
}

var store = &ModelRegistryStore{
	models:  make(map[string]*AIModel),
	metrics: make(map[string]*ModelMetrics),
}

func init() {
	seedModels()
}

func seedModels() {
	models := []*AIModel{
		{ID: "model-001", Name: "LFM2.5 Thinking", ModelID: "lfm2.5-thinking:1.2b", Category: "planner", SizeBytes: 700000000, Parameters: "1.2B", ContextLength: 4096, Capabilities: []string{"reasoning", "planning"}, Status: "active", MaxConcurrency: 10, CreatedAt: time.Now().Format(time.RFC3339)},
		{ID: "model-002", Name: "Command R7B", ModelID: "command-r7b:7b", Category: "customer", SizeBytes: 4000000000, Parameters: "7B", ContextLength: 128000, Capabilities: []string{"chat", "customer_support", "multi_turn"}, Status: "active", MaxConcurrency: 5, CreatedAt: time.Now().Format(time.RFC3339)},
		{ID: "model-003", Name: "Qwen2.5 Coder", ModelID: "qwen2.5-coder:3b", Category: "developer", SizeBytes: 2000000000, Parameters: "3B", ContextLength: 32000, Capabilities: []string{"code_generation", "code_review", "refactoring"}, Status: "active", MaxConcurrency: 8, CreatedAt: time.Now().Format(time.RFC3339)},
		{ID: "model-004", Name: "Qwen3 VL", ModelID: "qwen3-vl:4b", Category: "vision", SizeBytes: 3000000000, Parameters: "4B", ContextLength: 32000, Capabilities: []string{"image_analysis", "ocr", "multimodal"}, Status: "active", MaxConcurrency: 4, CreatedAt: time.Now().Format(time.RFC3339)},
		{ID: "model-005", Name: "WhiteRabbitNeo", ModelID: "whiterabbitneo:7b", Category: "security", SizeBytes: 4000000000, Parameters: "7B", ContextLength: 32000, Capabilities: []string{"security_analysis", "vulnerability_detection"}, Status: "active", MaxConcurrency: 3, CreatedAt: time.Now().Format(time.RFC3339)},
		{ID: "model-006", Name: "Llama 3.1", ModelID: "llama3.1:7b", Category: "business", SizeBytes: 4000000000, Parameters: "7B", ContextLength: 128000, Capabilities: []string{"analysis", "summarization", "extraction"}, Status: "active", MaxConcurrency: 6, CreatedAt: time.Now().Format(time.RFC3339)},
		{ID: "model-007", Name: "Phi4 Mini", ModelID: "phi4-mini:3.8b", Category: "rag", SizeBytes: 2500000000, Parameters: "3.8B", ContextLength: 16000, Capabilities: []string{"question_answering", "text_generation"}, Status: "active", MaxConcurrency: 10, CreatedAt: time.Now().Format(time.RFC3339)},
		{ID: "model-008", Name: "Nomic Embed Text", ModelID: "nomic-embed-text", Category: "embedding", SizeBytes: 274000000, Parameters: "137M", ContextLength: 8192, Capabilities: []string{"embeddings", "semantic_search"}, Status: "active", MaxConcurrency: 20, CreatedAt: time.Now().Format(time.RFC3339)},
	}

	store.mu.Lock()
	for _, m := range models {
		store.models[m.ID] = m
		store.metrics[m.ID] = &ModelMetrics{ModelID: m.ModelID, RequestCount: 0, AvgLatencyMs: 0, ErrorRate: 0, ActiveWorkers: 0}
	}
	store.mu.Unlock()
}

func main() {
	svcLogger := logger.New("model-registry")
	svcLogger.Info("Starting Model Registry Service")

	cfg, err := config.LoadConfig("")
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to load config: %v", err))
	}
	_ = cfg

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/models", listModelsHandler)
	mux.HandleFunc("/api/v1/models/create", createModelHandler)
	mux.HandleFunc("/api/v1/models/metrics/", metricsHandler)
	mux.HandleFunc("/api/v1/models/pull/", pullModelHandler)
	mux.HandleFunc("/api/v1/models/", modelHandler)

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

func listModelsHandler(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	models := make([]*AIModel, 0)
	for _, m := range store.models {
		models = append(models, m)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": models})
}

func createModelHandler(w http.ResponseWriter, r *http.Request) {
	var model AIModel
	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	model.ID = fmt.Sprintf("model-%d", time.Now().UnixNano())
	model.Status = "active"
	model.CreatedAt = time.Now().Format(time.RFC3339)

	store.mu.Lock()
	store.models[model.ID] = &model
	store.metrics[model.ID] = &ModelMetrics{ModelID: model.ModelID}
	store.mu.Unlock()

	writeJSON(w, http.StatusCreated, map[string]interface{}{"data": model})
}

func modelHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/models/"):]

	store.mu.RLock()
	model, exists := store.models[id]
	store.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Model not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": model})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	modelID := r.URL.Path[len("/api/v1/models/metrics/"):]

	store.mu.RLock()
	metric, exists := store.metrics[modelID]
	store.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Metrics not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": metric})
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
