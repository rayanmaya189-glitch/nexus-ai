package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

type VisionAnalysisRequest struct {
	ImageURL string `json:"image_url,omitempty"`
	ImageB64 string `json:"image_b64,omitempty"`
	Prompt   string `json:"prompt"`
	Model    string `json:"model,omitempty"`
}

type VisionAnalysisResponse struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Labels      []string `json:"labels"`
	Confidence  float64 `json:"confidence"`
	Model       string  `json:"model"`
	LatencyMs   float64 `json:"latency_ms"`
}

type OCRRequest struct {
	ImageURL string `json:"image_url,omitempty"`
	ImageB64 string `json:"image_b64,omitempty"`
	Language string `json:"language,omitempty"`
}

type OCRResponse struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	LatencyMs  float64 `json:"latency_ms"`
}

type ClassifyRequest struct {
	Text  string   `json:"text"`
	Labels []string `json:"labels,omitempty"`
}

type ClassifyResponse struct {
	ID         string             `json:"id"`
	Category   string             `json:"category"`
	Score      float64            `json:"score"`
	Scores     map[string]float64 `json:"scores"`
	LatencyMs  float64            `json:"latency_ms"`
}

type OllamaConfig struct {
	BaseURL string
	Timeout time.Duration
}

var ollamaCfg OllamaConfig
var httpClient *http.Client

func init() {
	cfg, _ := config.LoadConfig("")
	ollamaURL := getEnv("OLLAMA_BASE_URL", cfg.Ollama.BaseURL)
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	timeout := cfg.Ollama.Timeout
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	ollamaCfg = OllamaConfig{BaseURL: ollamaURL, Timeout: timeout}
	httpClient = &http.Client{Timeout: timeout}
}

func main() {
	svcLogger := logger.New("vision-service")
	svcLogger.Info("Starting Vision Service")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy", "service": "vision-service", "version": "1.0.0",
		})
	})

	mux.HandleFunc("/api/v1/vision/analyze", analyzeHandler)
	mux.HandleFunc("/api/v1/vision/ocr", ocrHandler)
	mux.HandleFunc("/api/v1/vision/classify", classifyHandler)

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8092")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Vision Service listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req VisionAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if req.Prompt == "" {
		req.Prompt = "Describe this image in detail"
	}

	model := req.Model
	if model == "" {
		model = "qwen3-vl:4b"
	}

	start := time.Now()

	prompt := fmt.Sprintf("Image analysis request. Model: %s. Prompt: %s. Image provided: %v",
		model, req.Prompt, req.ImageURL != "" || req.ImageB64 != "")

	ollamaReq := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	body, _ := json.Marshal(ollamaReq)
	resp, err := httpClient.Post(
		ollamaCfg.BaseURL+"/api/generate",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		writeError(w, http.StatusBadGateway, "AI_MODEL_ERROR", fmt.Sprintf("Ollama error: %v", err))
		return
	}
	defer resp.Body.Close()

	var ollamaResp struct {
		Response string `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&ollamaResp)

	latency := float64(time.Since(start).Milliseconds())

	result := VisionAnalysisResponse{
		ID:          generateID(),
		Description: ollamaResp.Response,
		Labels:      []string{"vision", "analysis"},
		Confidence:  0.85,
		Model:       model,
		LatencyMs:   latency,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func ocrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req OCRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	start := time.Now()

	prompt := "Extract all text from this image. Return only the text content."
	if req.Language != "" {
		prompt = fmt.Sprintf("Extract all text from this image in %s language. Return only the text content.", req.Language)
	}

	ollamaReq := map[string]interface{}{
		"model":  "qwen3-vl:4b",
		"prompt": prompt,
		"stream": false,
	}

	body, _ := json.Marshal(ollamaReq)
	resp, err := httpClient.Post(
		ollamaCfg.BaseURL+"/api/generate",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		writeError(w, http.StatusBadGateway, "AI_MODEL_ERROR", fmt.Sprintf("Ollama error: %v", err))
		return
	}
	defer resp.Body.Close()

	var ollamaResp struct {
		Response string `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&ollamaResp)

	latency := float64(time.Since(start).Milliseconds())

	result := OCRResponse{
		ID:         generateID(),
		Text:       ollamaResp.Response,
		Confidence: 0.90,
		LatencyMs:  latency,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func classifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req ClassifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "text is required")
		return
	}

	start := time.Now()

	labelsStr := "general categories"
	if len(req.Labels) > 0 {
		labelsStr = fmt.Sprintf("%v", req.Labels)
	}

	prompt := fmt.Sprintf("Classify the following text into one of these categories: %s. Text: %s. Return only the category name.", labelsStr, req.Text)

	ollamaReq := map[string]interface{}{
		"model":  "qwen2.5-coder:3b",
		"prompt": prompt,
		"stream": false,
	}

	body, _ := json.Marshal(ollamaReq)
	resp, err := httpClient.Post(
		ollamaCfg.BaseURL+"/api/generate",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		writeError(w, http.StatusBadGateway, "AI_MODEL_ERROR", fmt.Sprintf("Ollama error: %v", err))
		return
	}
	defer resp.Body.Close()

	var ollamaResp struct {
		Response string `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&ollamaResp)

	latency := float64(time.Since(start).Milliseconds())

	result := ClassifyResponse{
		ID:        generateID(),
		Category:  ollamaResp.Response,
		Score:     0.80,
		Scores:    map[string]float64{"primary": 0.80},
		LatencyMs: latency,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
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
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}


