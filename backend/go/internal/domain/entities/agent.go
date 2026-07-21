package entities

import "time"

type Agent struct {
	ID            int64     `json:"id" db:"id"`
	TenantID      int64     `json:"tenant_id" db:"tenant_id"`
	Name          string    `json:"name" db:"name"`
	AgentType     string    `json:"agent_type" db:"agent_type"`
	Model         string    `json:"model" db:"model"`
	SystemPrompt  string    `json:"system_prompt" db:"system_prompt"`
	Capabilities  string    `json:"capabilities" db:"capabilities"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type AgentExecution struct {
	ID          int64     `json:"id" db:"id"`
	AgentID     int64     `json:"agent_id" db:"agent_id"`
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Task        string    `json:"task" db:"task"`
	Status      string    `json:"status" db:"status"`
	Result      string    `json:"result" db:"result"`
	TokensUsed  int       `json:"tokens_used" db:"tokens_used"`
	LatencyMs   float64   `json:"latency_ms" db:"latency_ms"`
	StartedAt   time.Time `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}

type AgentStep struct {
	ID            int64     `json:"id" db:"id"`
	ExecutionID   int64     `json:"execution_id" db:"execution_id"`
	StepNumber    int       `json:"step_number" db:"step_number"`
	AgentType     string    `json:"agent_type" db:"agent_type"`
	Action        string    `json:"action" db:"action"`
	ToolName      string    `json:"tool_name" db:"tool_name"`
	ToolParams    string    `json:"tool_params" db:"tool_params"`
	Result        string    `json:"result" db:"result"`
	Status        string    `json:"status" db:"status"`
	StartedAt     time.Time `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
}
