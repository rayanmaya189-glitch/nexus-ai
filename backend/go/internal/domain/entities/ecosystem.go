package entities

import "time"

type Integration struct {
	ID              int64      `json:"id" db:"id"`
	TenantID        int64      `json:"tenant_id" db:"tenant_id"`
	Name            string     `json:"name" db:"name"`
	Type            string     `json:"type" db:"type"`
	Config          string     `json:"config" db:"config"`
	Status          string     `json:"status" db:"status"`
	CredentialsRef  string     `json:"credentials_ref" db:"credentials_ref"`
	LastHealthCheck *time.Time `json:"last_health_check" db:"last_health_check"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type MCPTool struct {
	ID            int64     `json:"id" db:"id"`
	TenantID      int64     `json:"tenant_id" db:"tenant_id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	InputSchema   string    `json:"input_schema" db:"input_schema"`
	IntegrationID *int64    `json:"integration_id" db:"integration_id"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type MCPToolInvocation struct {
	ID         int64     `json:"id" db:"id"`
	ToolID     int64     `json:"tool_id" db:"tool_id"`
	TenantID   int64     `json:"tenant_id" db:"tenant_id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Input      string    `json:"input" db:"input"`
	Output     string    `json:"output" db:"output"`
	Status     string    `json:"status" db:"status"`
	LatencyMs  float64   `json:"latency_ms" db:"latency_ms"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
