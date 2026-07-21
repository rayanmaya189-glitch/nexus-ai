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

type AuditLog struct {
	ID            string                 `json:"id"`
	TenantID      int64                  `json:"tenant_id"`
	UserID        int64                  `json:"user_id"`
	Action        string                 `json:"action"`
	ResourceType  string                 `json:"resource_type"`
	ResourceID    string                 `json:"resource_id"`
	Details       map[string]interface{} `json:"details"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	Status        string                 `json:"status"`
	CreatedAt     string                 `json:"created_at"`
}

type AuditQuery struct {
	TenantID     int64  `json:"tenant_id"`
	UserID       *int64 `json:"user_id,omitempty"`
	Action       string `json:"action,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
	Limit        int    `json:"limit,omitempty"`
}

type AuditStats struct {
	TotalEvents    int64             `json:"total_events"`
	ByAction       map[string]int64  `json:"by_action"`
	ByResource     map[string]int64  `json:"by_resource"`
	ByStatus       map[string]int64  `json:"by_status"`
	RecentActivity []AuditLog        `json:"recent_activity"`
}

type AuditStore struct {
	logs []AuditLog
	mu   sync.RWMutex
}

var store = &AuditStore{logs: make([]AuditLog, 0)}

func init() {
	seedAuditLogs()
}

func seedAuditLogs() {
	sampleLogs := []AuditLog{
		{ID: "audit-001", TenantID: 1, UserID: 1, Action: "login.success", ResourceType: "user", ResourceID: "1", Details: map[string]interface{}{"method": "password"}, IPAddress: "192.168.1.1", Status: "success", CreatedAt: time.Now().Add(-2 * time.Hour).Format(time.RFC3339)},
		{ID: "audit-002", TenantID: 1, UserID: 1, Action: "agent.execute", ResourceType: "agent", ResourceID: "agent-001", Details: map[string]interface{}{"task": "customer inquiry"}, IPAddress: "192.168.1.1", Status: "success", CreatedAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339)},
		{ID: "audit-003", TenantID: 1, UserID: 2, Action: "document.ingest", ResourceType: "document", ResourceID: "doc-001", Details: map[string]interface{}{"chunks": 15}, IPAddress: "10.0.0.1", Status: "success", CreatedAt: time.Now().Add(-30 * time.Minute).Format(time.RFC3339)},
		{ID: "audit-004", TenantID: 1, UserID: 1, Action: "user.create", ResourceType: "user", ResourceID: "3", Details: map[string]interface{}{"email": "new@example.com"}, IPAddress: "192.168.1.1", Status: "success", CreatedAt: time.Now().Add(-15 * time.Minute).Format(time.RFC3339)},
		{ID: "audit-005", TenantID: 2, UserID: 5, Action: "api_key.create", ResourceType: "api_key", ResourceID: "key-001", Details: map[string]interface{}{"permissions": []string{"read"}}, IPAddress: "172.16.0.1", Status: "success", CreatedAt: time.Now().Add(-5 * time.Minute).Format(time.RFC3339)},
	}

	store.mu.Lock()
	store.logs = append(store.logs, sampleLogs...)
	store.mu.Unlock()
}

func main() {
	log.Println("Starting Audit Service")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/audit/logs", logsHandler)
	mux.HandleFunc("/api/v1/audit/logs/create", createLogHandler)
	mux.HandleFunc("/api/v1/audit/stats", statsHandler)
	mux.HandleFunc("/api/v1/audit/search", searchHandler)

	port := getEnv("PORT", "8085")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Audit Service listening on %s", addr)

	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "healthy", "service": "audit-service", "version": "1.0.0",
	})
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": store.logs})
}

func createLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var logEntry AuditLog
	if err := json.NewDecoder(r.Body).Decode(&logEntry); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	logEntry.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	if logEntry.CreatedAt == "" {
		logEntry.CreatedAt = time.Now().Format(time.RFC3339)
	}
	if logEntry.Status == "" {
		logEntry.Status = "success"
	}

	store.mu.Lock()
	store.logs = append(store.logs, logEntry)
	store.mu.Unlock()

	writeJSON(w, http.StatusCreated, map[string]interface{}{"data": logEntry})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	byAction := make(map[string]int64)
	byResource := make(map[string]int64)
	byStatus := make(map[string]int64)

	for _, l := range store.logs {
		byAction[l.Action]++
		byResource[l.ResourceType]++
		byStatus[l.Status]++
	}

	recent := make([]AuditLog, 0)
	start := len(store.logs) - 5
	if start < 0 {
		start = 0
	}
	recent = append(recent, store.logs[start:]...)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": AuditStats{
			TotalEvents:    int64(len(store.logs)),
			ByAction:       byAction,
			ByResource:     byResource,
			ByStatus:       byStatus,
			RecentActivity: recent,
		},
	})
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var query AuditQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	results := make([]AuditLog, 0)
	for _, l := range store.logs {
		if l.TenantID != query.TenantID && query.TenantID != 0 {
			continue
		}
		if query.UserID != nil && l.UserID != *query.UserID {
			continue
		}
		if query.Action != "" && l.Action != query.Action {
			continue
		}
		if query.ResourceType != "" && l.ResourceType != query.ResourceType {
			continue
		}
		results = append(results, l)
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}
	if len(results) > limit {
		results = results[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"logs":  results,
			"total": len(results),
		},
	})
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
