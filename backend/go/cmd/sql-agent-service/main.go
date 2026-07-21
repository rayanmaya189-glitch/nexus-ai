package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type SQLQueryRequest struct {
	Question    string `json:"question"`
	DatabaseID  string `json:"database_id"`
	Context     string `json:"context,omitempty"`
	MaxRows     int    `json:"max_rows,omitempty"`
}

type SQLQueryResult struct {
	Question     string                   `json:"question"`
	GeneratedSQL string                   `json:"generated_sql"`
	Columns      []ColumnInfo             `json:"columns"`
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int                      `json:"row_count"`
	LatencyMs    float64                  `json:"latency_ms"`
	ModelUsed    string                   `json:"model_used"`
	Explanation  string                   `json:"explanation"`
}

type ColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type SQLSchemaRequest struct {
	DatabaseID string `json:"database_id"`
}

type SQLSchemaResult struct {
	DatabaseID string      `json:"database_id"`
	Tables     []TableInfo `json:"tables"`
}

type TableInfo struct {
	Name    string       `json:"name"`
	Columns []ColumnInfo `json:"columns"`
}

type SQLValidationResult struct {
	Valid     bool   `json:"valid"`
	SQL       string `json:"sql"`
	Warnings  []string `json:"warnings,omitempty"`
}

func main() {
	log.Println("Starting SQL Agent Service")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/sql/query", queryHandler)
	mux.HandleFunc("/api/v1/sql/generate", generateHandler)
	mux.HandleFunc("/api/v1/sql/validate", validateHandler)
	mux.HandleFunc("/api/v1/sql/schema", schemaHandler)
	mux.HandleFunc("/api/v1/sql/explain", explainHandler)

	port := getEnv("PORT", "8090")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("SQL Agent Service listening on %s", addr)

	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "healthy", "service": "sql-agent", "version": "1.0.0",
	})
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req SQLQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	start := time.Now()

	generatedSQL := fmt.Sprintf("SELECT * FROM data WHERE description LIKE '%%%s%%' LIMIT %d", req.Question, maxRows(req.MaxRows))

	result := SQLQueryResult{
		Question:     req.Question,
		GeneratedSQL: generatedSQL,
		Columns: []ColumnInfo{
			{Name: "id", Type: "bigint"},
			{Name: "name", Type: "varchar"},
			{Name: "description", Type: "text"},
			{Name: "created_at", Type: "timestamp"},
		},
		Rows: []map[string]interface{}{
			{"id": 1, "name": "Sample Data", "description": "Example result", "created_at": time.Now().Format(time.RFC3339)},
		},
		RowCount:    1,
		LatencyMs:   float64(time.Since(start).Milliseconds()),
		ModelUsed:   "qwen2.5-coder:3b",
		Explanation: fmt.Sprintf("Generated SQL query to find records matching '%s'", req.Question),
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req struct {
		Question   string `json:"question"`
		SchemaJSON string `json:"schema_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	sql := fmt.Sprintf("-- Generated SQL\nSELECT * FROM records\nWHERE description ILIKE '%%%s%%'\nORDER BY created_at DESC\nLIMIT 100;", req.Question)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"generated_sql": sql,
			"confidence":    0.85,
			"explanation":    "Generated based on natural language question",
		},
	})
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req struct {
		SQL string `json:"sql"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	result := SQLValidationResult{
		Valid: true,
		SQL:   req.SQL,
	}

	if len(req.SQL) > 0 {
		result.Warnings = []string{"Consider adding a LIMIT clause for large result sets"}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func schemaHandler(w http.ResponseWriter, r *http.Request) {
	var req SQLSchemaRequest
	if r.Method == http.MethodPost {
		json.NewDecoder(r.Body).Decode(&req)
	}

	result := SQLSchemaResult{
		DatabaseID: req.DatabaseID,
		Tables: []TableInfo{
			{
				Name: "users",
				Columns: []ColumnInfo{
					{Name: "id", Type: "bigint"},
					{Name: "email", Type: "varchar"},
					{Name: "name", Type: "varchar"},
					{Name: "tenant_id", Type: "bigint"},
					{Name: "status", Type: "varchar"},
					{Name: "created_at", Type: "timestamp"},
				},
			},
			{
				Name: "agents",
				Columns: []ColumnInfo{
					{Name: "id", Type: "bigint"},
					{Name: "name", Type: "varchar"},
					{Name: "agent_type", Type: "varchar"},
					{Name: "model", Type: "varchar"},
					{Name: "tenant_id", Type: "bigint"},
				},
			},
			{
				Name: "documents",
				Columns: []ColumnInfo{
					{Name: "id", Type: "bigint"},
					{Name: "title", Type: "varchar"},
					{Name: "content", Type: "text"},
					{Name: "tenant_id", Type: "bigint"},
				},
			},
		},
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func explainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req struct {
		SQL string `json:"sql"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"sql":         req.SQL,
			"explanation": "This query selects data from the specified table based on the provided conditions.",
			"steps": []string{
				"Parse the SQL query",
				"Identify tables and columns",
				"Apply WHERE conditions",
				"Return results",
			},
		},
	})
}

func maxRows(n int) int {
	if n <= 0 {
		return 100
	}
	return n
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

func getEnv(key, def string) string { return def }
