package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

type SecurityPolicy struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	PolicyType  string    `json:"policy_type"`
	Rules       []string  `json:"rules"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type FilterRequest struct {
	Content   string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
	TenantID  int64  `json:"tenant_id,omitempty"`
}

type FilterResponse struct {
	ID         string   `json:"id"`
	Safe       bool     `json:"safe"`
	Violations []string `json:"violations,omitempty"`
	Severity   string   `json:"severity,omitempty"`
	LatencyMs  float64  `json:"latency_ms"`
}

type ScanRequest struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
}

type ScanResponse struct {
	ID            string   `json:"id"`
	Vulnerable    bool     `json:"vulnerable"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
	LatencyMs     float64  `json:"latency_ms"`
}

type Vulnerability struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location,omitempty"`
}

type PolicyStore struct {
	policies map[int64]*SecurityPolicy
	nextID   int64
	mu       sync.RWMutex
}

var store = &PolicyStore{
	policies: make(map[int64]*SecurityPolicy),
	nextID:   1,
}

func init() {
	defaultPolicies := []*SecurityPolicy{
		{ID: 1, Name: "SQL Injection Prevention", PolicyType: "injection", Rules: []string{"DROP TABLE", "DELETE FROM", "TRUNCATE", "1=1", "OR 1=1"}, Severity: "critical", Status: "active", CreatedAt: time.Now()},
		{ID: 2, Name: "XSS Prevention", PolicyType: "xss", Rules: []string{"<script", "javascript:", "onerror=", "onload="}, Severity: "high", Status: "active", CreatedAt: time.Now()},
		{ID: 3, Name: "Prompt Injection Prevention", PolicyType: "prompt_injection", Rules: []string{"ignore previous", "ignore all", "disregard instructions", "you are now"}, Severity: "critical", Status: "active", CreatedAt: time.Now()},
		{ID: 4, Name: "Sensitive Data Exposure", PolicyType: "data_exposure", Rules: []string{"password", "secret", "api_key", "private_key"}, Severity: "medium", Status: "active", CreatedAt: time.Now()},
	}
	for _, p := range defaultPolicies {
		store.policies[p.ID] = p
	}
	store.nextID = 5
}

func main() {
	svcLogger := logger.New("security-service")
	svcLogger.Info("Starting Security Service")

	cfg, err := config.LoadConfig("")
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to load config: %v", err))
	}
	_ = cfg

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy", "service": "security-service", "version": "1.0.0",
		})
	})

	mux.HandleFunc("/api/v1/security/filter", filterHandler)
	mux.HandleFunc("/api/v1/security/scan", scanHandler)
	mux.HandleFunc("/api/v1/security/policies", policiesHandler)

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8094")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Security Service listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func filterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "content is required")
		return
	}

	start := time.Now()
	contentLower := strings.ToLower(req.Content)

	store.mu.RLock()
	defer store.mu.RUnlock()

	var violations []string
	var maxSeverity string

	for _, policy := range store.policies {
		if policy.Status != "active" {
			continue
		}
		for _, rule := range policy.Rules {
			if strings.Contains(contentLower, strings.ToLower(rule)) {
				violations = append(violations, fmt.Sprintf("%s: %s", policy.Name, rule))
				if severityRank(policy.Severity) > severityRank(maxSeverity) {
					maxSeverity = policy.Severity
				}
			}
		}
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0

	result := FilterResponse{
		ID:         generateID(),
		Safe:       len(violations) == 0,
		Violations: violations,
		Severity:   maxSeverity,
		LatencyMs:  latency,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func scanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "content is required")
		return
	}

	start := time.Now()
	contentLower := strings.ToLower(req.Content)

	var vulns []Vulnerability

	if strings.Contains(contentLower, "union select") || strings.Contains(contentLower, "drop table") {
		vulns = append(vulns, Vulnerability{
			Type:        "sql_injection",
			Severity:    "critical",
			Description: "Potential SQL injection detected",
		})
	}

	if strings.Contains(contentLower, "<script") || strings.Contains(contentLower, "javascript:") {
		vulns = append(vulns, Vulnerability{
			Type:        "xss",
			Severity:    "high",
			Description: "Potential cross-site scripting (XSS) detected",
		})
	}

	if strings.Contains(contentLower, "ignore previous") || strings.Contains(contentLower, "disregard") {
		vulns = append(vulns, Vulnerability{
			Type:        "prompt_injection",
			Severity:    "critical",
			Description: "Potential prompt injection attack detected",
		})
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0

	result := ScanResponse{
		ID:              generateID(),
		Vulnerable:      len(vulns) > 0,
		Vulnerabilities: vulns,
		LatencyMs:       latency,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func policiesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET allowed")
		return
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	policies := make([]*SecurityPolicy, 0, len(store.policies))
	for _, p := range store.policies {
		policies = append(policies, p)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": policies})
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
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
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
