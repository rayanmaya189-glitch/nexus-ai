package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Model       string        `json:"model,omitempty"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	AgentID     string        `json:"agent_id,omitempty"`
	SessionID   string        `json:"session_id,omitempty"`
}

type ChatResponse struct {
	ID        string        `json:"id"`
	Content   string        `json:"content"`
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	Tokens    TokenUsage    `json:"tokens"`
	LatencyMs float64       `json:"latency_ms"`
	CreatedAt string        `json:"created_at"`
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type StreamChunk struct {
	Delta     string `json:"delta"`
	Finished  bool   `json:"finished"`
	RequestID string `json:"request_id"`
}

type AIServiceConfig struct {
	OllamaURL string
	Timeout   time.Duration
	Models    map[string]string
}

type AIGateway struct {
	config     *AIServiceConfig
	httpClient *http.Client
	sessions   map[string]*ChatSession
	mu         sync.RWMutex
}

type ChatSession struct {
	ID        string        `json:"id"`
	Messages  []ChatMessage `json:"messages"`
	Model     string        `json:"model"`
	CreatedAt time.Time     `json:"created_at"`
}

var gateway *AIGateway

func init() {
	cfg, _ := config.LoadConfig("")

	ollamaURL := getEnv("OLLAMA_BASE_URL", cfg.Ollama.BaseURL)
	ollamaTimeout := cfg.Ollama.Timeout
	if ollamaTimeout == 0 {
		ollamaTimeout = 120 * time.Second
	}

	aiConfig := &AIServiceConfig{
		OllamaURL: ollamaURL,
		Timeout:   ollamaTimeout,
		Models: map[string]string{
			"planner":  "lfm2.5-thinking:1.2b",
			"customer": "command-r7b:7b",
			"developer": "qwen2.5-coder:3b",
			"vision":   "qwen3-vl:4b",
			"security": "whiterabbitneo:7b",
			"business": "llama3.1:7b",
			"rag":      "phi4-mini:3.8b",
			"sql":      "qwen2.5-coder:3b",
		},
	}

	gateway = &AIGateway{
		config: aiConfig,
		httpClient: &http.Client{
			Timeout: ollamaTimeout,
		},
		sessions: make(map[string]*ChatSession),
	}
}

func main() {
	svcLogger := logger.New("ai-gateway")
	svcLogger.Info("Starting AI Gateway Service")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/ai/chat", chatHandler)
	mux.HandleFunc("/api/v1/ai/chat/stream", chatStreamHandler)
	mux.HandleFunc("/api/v1/ai/sessions", sessionsHandler)
	mux.HandleFunc("/api/v1/ai/sessions/", sessionHandler)
	mux.HandleFunc("/api/v1/ai/models", modelsHandler)
	mux.HandleFunc("/api/v1/ai/completions", completionsHandler)

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8080")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("AI Gateway listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "ai-gateway",
		"version": "1.0.0",
	})
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "messages is required")
		return
	}

	model := req.Model
	if model == "" {
		model = gateway.config.Models["rag"]
	}
	if agentModel, ok := gateway.config.Models[req.AgentID]; ok {
		model = agentModel
	}

	start := time.Now()

	prompt := buildPrompt(req.Messages)
	ollamaReq := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	body, _ := json.Marshal(ollamaReq)
	resp, err := gateway.httpClient.Post(
		gateway.config.OllamaURL+"/api/generate",
		"application/json",
		bytesReader(body),
	)
	if err != nil {
		writeError(w, http.StatusBadGateway, "AI_MODEL_ERROR", fmt.Sprintf("Ollama error: %v", err))
		return
	}
	defer resp.Body.Close()

	var ollamaResp struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}
	json.NewDecoder(resp.Body).Decode(&ollamaResp)

	latency := float64(time.Since(start).Milliseconds())

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = generateID()
	}

	gateway.mu.Lock()
	session, exists := gateway.sessions[sessionID]
	if !exists {
		session = &ChatSession{
			ID:        sessionID,
			Messages:  []ChatMessage{},
			Model:     model,
			CreatedAt: time.Now(),
		}
		gateway.sessions[sessionID] = session
	}
	session.Messages = append(session.Messages, req.Messages...)
	session.Messages = append(session.Messages, ChatMessage{Role: "assistant", Content: ollamaResp.Response})
	gateway.mu.Unlock()

	resp2 := ChatResponse{
		ID:        sessionID,
		Content:   ollamaResp.Response,
		Model:     model,
		Messages:  session.Messages,
		Tokens: TokenUsage{
			PromptTokens:     estimateTokens(prompt),
			CompletionTokens: estimateTokens(ollamaResp.Response),
			TotalTokens:      estimateTokens(prompt) + estimateTokens(ollamaResp.Response),
		},
		LatencyMs: latency,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": resp2})
}

func chatStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	model := req.Model
	if model == "" {
		model = gateway.config.Models["rag"]
	}
	if agentModel, ok := gateway.config.Models[req.AgentID]; ok {
		model = agentModel
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "STREAM_ERROR", "Streaming not supported")
		return
	}

	prompt := buildPrompt(req.Messages)
	ollamaReq := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": true,
	}

	body, _ := json.Marshal(ollamaReq)
	resp, err := gateway.httpClient.Post(
		gateway.config.OllamaURL+"/api/generate",
		"application/json",
		bytesReader(body),
	)
	if err != nil {
		writeError(w, http.StatusBadGateway, "AI_MODEL_ERROR", err.Error())
		return
	}
	defer resp.Body.Close()

	requestID := generateID()
	decoder := json.NewDecoder(resp.Body)

	for {
		var chunk struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := decoder.Decode(&chunk); err != nil {
			break
		}

		sseChunk := StreamChunk{
			Delta:     chunk.Response,
			Finished:  chunk.Done,
			RequestID: requestID,
		}

		chunkJSON, _ := json.Marshal(sseChunk)
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()

		if chunk.Done {
			break
		}
	}
}

func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	gateway.mu.RLock()
	defer gateway.mu.RUnlock()

	sessions := make([]ChatSession, 0)
	for _, s := range gateway.sessions {
		sessions = append(sessions, *s)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": sessions})
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Path[len("/api/v1/ai/sessions/"):]

	gateway.mu.RLock()
	session, exists := gateway.sessions[sessionID]
	gateway.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Session not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": session})
}

func modelsHandler(w http.ResponseWriter, r *http.Request) {
	type ModelInfo struct {
		Name     string `json:"name"`
		ModelID  string `json:"model_id"`
		Category string `json:"category"`
	}

	models := []ModelInfo{
		{Name: "LFM2.5 Thinking", ModelID: "lfm2.5-thinking:1.2b", Category: "planner"},
		{Name: "Command R7B", ModelID: "command-r7b:7b", Category: "customer"},
		{Name: "Qwen2.5 Coder", ModelID: "qwen2.5-coder:3b", Category: "developer"},
		{Name: "Qwen3 VL", ModelID: "qwen3-vl:4b", Category: "vision"},
		{Name: "WhiteRabbitNeo", ModelID: "whiterabbitneo:7b", Category: "security"},
		{Name: "Llama 3.1", ModelID: "llama3.1:7b", Category: "business"},
		{Name: "Phi4 Mini", ModelID: "phi4-mini:3.8b", Category: "rag"},
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": models})
}

func completionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req struct {
		Prompt      string  `json:"prompt"`
		Model       string  `json:"model"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"max_tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	start := time.Now()

	ollamaReq := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt,
		"stream": false,
	}

	body, _ := json.Marshal(ollamaReq)
	resp, err := gateway.httpClient.Post(
		gateway.config.OllamaURL+"/api/generate",
		"application/json",
		bytesReader(body),
	)
	if err != nil {
		writeError(w, http.StatusBadGateway, "AI_MODEL_ERROR", err.Error())
		return
	}
	defer resp.Body.Close()

	var ollamaResp struct {
		Response string `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&ollamaResp)

	latency := float64(time.Since(start).Milliseconds())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"text":       ollamaResp.Response,
			"model":      req.Model,
			"latency_ms": latency,
			"tokens": TokenUsage{
				PromptTokens:     estimateTokens(req.Prompt),
				CompletionTokens: estimateTokens(ollamaResp.Response),
				TotalTokens:      estimateTokens(req.Prompt) + estimateTokens(ollamaResp.Response),
			},
		},
	})
}

func buildPrompt(messages []ChatMessage) string {
	if len(messages) == 1 {
		return messages[0].Content
	}
	var result string
	for _, m := range messages {
		result += fmt.Sprintf("[%s]: %s\n", m.Role, m.Content)
	}
	return result
}

func estimateTokens(text string) int {
	return len(text) / 4
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func bytesReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}
