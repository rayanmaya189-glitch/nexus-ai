package queries

type GetAuditLogQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListAuditLogsQuery struct {
	TenantID int64 `json:"tenant_id"`
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
}

type SearchAuditLogsQuery struct {
	TenantID     int64  `json:"tenant_id"`
	Query        string `json:"query"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	Status       string `json:"status"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Page         int    `json:"page"`
	PerPage      int    `json:"per_page"`
}
