package queries

type GetAgentQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListAgentsQuery struct {
	TenantID  int64  `json:"tenant_id"`
	AgentType string `json:"agent_type"`
}

type GetAgentExecutionQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListAgentExecutionsQuery struct {
	AgentID int64 `json:"agent_id"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
}
