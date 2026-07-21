package queries

type GetWorkflowQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListWorkflowsQuery struct {
	TenantID int64 `json:"tenant_id"`
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
}

type GetWorkflowExecutionQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListWorkflowExecutionsQuery struct {
	WorkflowID int64 `json:"workflow_id"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
}
