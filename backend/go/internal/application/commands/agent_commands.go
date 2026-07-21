package commands

type CreateAgentCommand struct {
	TenantID     int64  `json:"tenant_id" validate:"required"`
	Name         string `json:"name" validate:"required"`
	AgentType    string `json:"agent_type" validate:"required"`
	Model        string `json:"model" validate:"required"`
	SystemPrompt string `json:"system_prompt"`
	Capabilities string `json:"capabilities"`
}

type UpdateAgentCommand struct {
	ID           int64  `json:"id" validate:"required"`
	Name         string `json:"name"`
	AgentType    string `json:"agent_type"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt"`
	Capabilities string `json:"capabilities"`
	Status       string `json:"status"`
}

type ExecuteAgentTaskCommand struct {
	AgentID   int64             `json:"agent_id" validate:"required"`
	TenantID  int64             `json:"tenant_id" validate:"required"`
	UserID    int64             `json:"user_id" validate:"required"`
	Task      string            `json:"task" validate:"required"`
	Context   string            `json:"context"`
	Metadata  map[string]string `json:"metadata"`
}
