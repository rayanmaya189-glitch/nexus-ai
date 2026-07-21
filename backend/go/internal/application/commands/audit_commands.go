package commands

type CreateAuditLogCommand struct {
	TenantID     int64  `json:"tenant_id" validate:"required"`
	UserID       int64  `json:"user_id" validate:"required"`
	Action       string `json:"action" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	ResourceID   string `json:"resource_id"`
	Details      string `json:"details"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	Status       string `json:"status"`
}
