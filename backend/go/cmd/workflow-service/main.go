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

type Workflow struct {
	ID          string         `json:"id"`
	TenantID    int64          `json:"tenant_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	TriggerType string         `json:"trigger_type"`
	Steps       []WorkflowStep `json:"steps"`
	Status      string         `json:"status"`
	Version     int            `json:"version"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

type WorkflowStep struct {
	ID         string                 `json:"id"`
	StepNumber int                    `json:"step_number"`
	Name       string                 `json:"name"`
	AgentType  string                 `json:"agent_type"`
	Action     string                 `json:"action"`
	Config     map[string]interface{} `json:"config"`
	Timeout    int                    `json:"timeout_seconds"`
	RetryPolicy *RetryPolicy          `json:"retry_policy,omitempty"`
}

type RetryPolicy struct {
	MaxRetries    int `json:"max_retries"`
	BackoffMs     int `json:"backoff_ms"`
}

type WorkflowExecution struct {
	ID          string         `json:"id"`
	WorkflowID  string         `json:"workflow_id"`
	TenantID    int64          `json:"tenant_id"`
	UserID      int64          `json:"user_id"`
	Status      string         `json:"status"`
	StepResults []StepResult   `json:"step_results"`
	Input       interface{}    `json:"input"`
	Output      interface{}    `json:"output,omitempty"`
	Error       string         `json:"error,omitempty"`
	StartedAt   string         `json:"started_at"`
	CompletedAt string         `json:"completed_at,omitempty"`
}

type StepResult struct {
	StepID    string      `json:"step_id"`
	Status    string      `json:"status"`
	Output    interface{} `json:"output,omitempty"`
	Error     string      `json:"error,omitempty"`
	StartedAt string      `json:"started_at"`
	EndedAt   string      `json:"ended_at,omitempty"`
}

type WorkflowStore struct {
	workflows   map[string]*Workflow
	executions  map[string]*WorkflowExecution
	mu          sync.RWMutex
}

var store = &WorkflowStore{
	workflows:  make(map[string]*Workflow),
	executions: make(map[string]*WorkflowExecution),
}

func init() {
	defaultWorkflows()
}

func defaultWorkflows() {
	seedWorkflow := &Workflow{
		ID:          "wf-001",
		TenantID:    1,
		Name:        "Customer Support Pipeline",
		Description: "Automated customer support workflow with escalation",
		TriggerType: "chat_message",
		Steps: []WorkflowStep{
			{
				ID: "step-1", StepNumber: 1, Name: "Classify Intent",
				AgentType: "classifier", Action: "classify_intent",
				Config: map[string]interface{}{"categories": []string{"billing", "technical", "general"}},
				Timeout: 30,
			},
			{
				ID: "step-2", StepNumber: 2, Name: "Route to Agent",
				AgentType: "router", Action: "route_to_specialist",
				Config: map[string]interface{}{"routing_rules": "intent_based"},
				Timeout: 10,
			},
			{
				ID: "step-3", StepNumber: 3, Name: "Generate Response",
				AgentType: "customer", Action: "generate_response",
				Config: map[string]interface{}{"model": "command-r7b:7b"},
				Timeout: 60,
				RetryPolicy: &RetryPolicy{MaxRetries: 2, BackoffMs: 1000},
			},
		},
		Status:    "active",
		Version:   1,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	store.mu.Lock()
	store.workflows[seedWorkflow.ID] = seedWorkflow
	store.mu.Unlock()
}

func main() {
	log.Println("Starting Workflow Service")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/workflows", listWorkflowsHandler)
	mux.HandleFunc("/api/v1/workflows/create", createWorkflowHandler)
	mux.HandleFunc("/api/v1/workflows/", workflowHandler)
	mux.HandleFunc("/api/v1/workflows/execute/", executeWorkflowHandler)
	mux.HandleFunc("/api/v1/workflows/executions/", getExecutionHandler)

	port := getEnv("PORT", "8084")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Workflow Service listening on %s", addr)

	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "healthy", "service": "workflow-service", "version": "1.0.0",
	})
}

func listWorkflowsHandler(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	workflows := make([]*Workflow, 0)
	for _, wf := range store.workflows {
		workflows = append(workflows, wf)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": workflows})
}

func createWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var wf Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	wf.ID = fmt.Sprintf("wf-%d", time.Now().UnixNano())
	wf.Status = "active"
	wf.Version = 1
	wf.CreatedAt = time.Now().Format(time.RFC3339)
	wf.UpdatedAt = time.Now().Format(time.RFC3339)

	for i := range wf.Steps {
		wf.Steps[i].ID = fmt.Sprintf("step-%d", i+1)
		wf.Steps[i].StepNumber = i + 1
	}

	store.mu.Lock()
	store.workflows[wf.ID] = &wf
	store.mu.Unlock()

	writeJSON(w, http.StatusCreated, map[string]interface{}{"data": wf})
}

func workflowHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/workflows/"):]

	store.mu.RLock()
	wf, exists := store.workflows[id]
	store.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Workflow not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": wf})
}

func executeWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	workflowID := r.URL.Path[len("/api/v1/workflows/execute/"):]

	store.mu.RLock()
	wf, exists := store.workflows[workflowID]
	store.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Workflow not found")
		return
	}

	var input interface{}
	json.NewDecoder(r.Body).Decode(&input)

	execID := fmt.Sprintf("exec-%d", time.Now().UnixNano())
	execution := &WorkflowExecution{
		ID:         execID,
		WorkflowID: workflowID,
		Status:     "running",
		Input:      input,
		StartedAt:  time.Now().Format(time.RFC3339),
	}

	for _, step := range wf.Steps {
		stepResult := StepResult{
			StepID:    step.ID,
			Status:    "completed",
			StartedAt: time.Now().Format(time.RFC3339),
			EndedAt:   time.Now().Format(time.RFC3339),
			Output: map[string]string{
				"message": fmt.Sprintf("Step '%s' completed", step.Name),
			},
		}
		execution.StepResults = append(execution.StepResults, stepResult)
	}

	execution.Status = "completed"
	completedAt := time.Now().Format(time.RFC3339)
	execution.CompletedAt = completedAt
	execution.Output = map[string]string{"result": "Workflow completed successfully"}

	store.mu.Lock()
	store.executions[execID] = execution
	store.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": execution})
}

func getExecutionHandler(w http.ResponseWriter, r *http.Request) {
	execID := r.URL.Path[len("/api/v1/workflows/executions/"):]

	store.mu.RLock()
	exec, exists := store.executions[execID]
	store.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Execution not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": exec})
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
