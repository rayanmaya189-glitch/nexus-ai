package commands

type CreateWorkflowCommand struct {
	TenantID    int64  `json:"tenant_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	TriggerType string `json:"trigger_type"`
	Config      string `json:"config"`
}

type UpdateWorkflowCommand struct {
	ID          int64  `json:"id" validate:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TriggerType string `json:"trigger_type"`
	Config      string `json:"config"`
	Status      string `json:"status"`
}

type ExecuteWorkflowCommand struct {
	WorkflowID int64             `json:"workflow_id" validate:"required"`
	TenantID   int64             `json:"tenant_id" validate:"required"`
	UserID     int64             `json:"user_id" validate:"required"`
	Input      string            `json:"input"`
	Context    map[string]string `json:"context"`
}
