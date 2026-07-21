package entities

import "time"

type AISession struct {
	ID        int64     `json:"id" db:"id"`
	TenantID  int64     `json:"tenant_id" db:"tenant_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	AgentID   int64     `json:"agent_id" db:"agent_id"`
	Model     string    `json:"model" db:"model"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type AIRequest struct {
	ID             int64     `json:"id" db:"id"`
	SessionID      int64     `json:"session_id" db:"session_id"`
	TenantID       int64     `json:"tenant_id" db:"tenant_id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	AgentID        int64     `json:"agent_id" db:"agent_id"`
	Prompt         string    `json:"prompt" db:"prompt"`
	Response       string    `json:"response" db:"response"`
	Model          string    `json:"model" db:"model"`
	PromptTokens   int       `json:"prompt_tokens" db:"prompt_tokens"`
	ResponseTokens int       `json:"response_tokens" db:"response_tokens"`
	LatencyMs      float64   `json:"latency_ms" db:"latency_ms"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type ChatMessage struct {
	ID        int64     `json:"id" db:"id"`
	SessionID int64     `json:"session_id" db:"session_id"`
	Role      string    `json:"role" db:"role"`
	Content   string    `json:"content" db:"content"`
	Tokens    int       `json:"tokens" db:"tokens"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Memory struct {
	ID           int64     `json:"id" db:"id"`
	TenantID     int64     `json:"tenant_id" db:"tenant_id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	AgentID      int64     `json:"agent_id" db:"agent_id"`
	SessionID    string    `json:"session_id" db:"session_id"`
	MemoryType   string    `json:"memory_type" db:"memory_type"`
	Content      string    `json:"content" db:"content"`
	Summary      string    `json:"summary" db:"summary"`
	Importance   float64   `json:"importance" db:"importance"`
	AccessCount  int       `json:"access_count" db:"access_count"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Workflow struct {
	ID          int64     `json:"id" db:"id"`
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	TriggerType string    `json:"trigger_type" db:"trigger_type"`
	Config      string    `json:"config" db:"config"`
	Status      string    `json:"status" db:"status"`
	Version     int       `json:"version" db:"version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type WorkflowStep struct {
	ID          int64     `json:"id" db:"id"`
	WorkflowID  int64     `json:"workflow_id" db:"workflow_id"`
	StepNumber  int       `json:"step_number" db:"step_number"`
	Name        string    `json:"name" db:"name"`
	AgentType   string    `json:"agent_type" db:"agent_type"`
	Action      string    `json:"action" db:"action"`
	Config      string    `json:"config" db:"config"`
	Timeout     int       `json:"timeout" db:"timeout"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type AIModel struct {
	ID             int64     `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	ModelID        string    `json:"model_id" db:"model_id"`
	Category       string    `json:"category" db:"category"`
	SizeBytes      int64     `json:"size_bytes" db:"size_bytes"`
	Parameters     string    `json:"parameters" db:"parameters"`
	ContextLength  int       `json:"context_length" db:"context_length"`
	Capabilities   string    `json:"capabilities" db:"capabilities"`
	Status         string    `json:"status" db:"status"`
	MaxConcurrency int       `json:"max_concurrency" db:"max_concurrency"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type AuditLog struct {
	ID           int64     `json:"id" db:"id"`
	TenantID     int64     `json:"tenant_id" db:"tenant_id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	Details      string    `json:"details" db:"details"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
