package queries

type GetUserQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListUsersQuery struct {
	TenantID int64 `json:"tenant_id"`
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
}

type GetTenantQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListTenantsQuery struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type GetRoleQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListRolesQuery struct {
	TenantID int64 `json:"tenant_id"`
}

type CheckPermissionQuery struct {
	UserID     int64  `json:"user_id" validate:"required"`
	Resource   string `json:"resource" validate:"required"`
	Action     string `json:"action" validate:"required"`
}

type GetUserByEmailQuery struct {
	Email string `json:"email" validate:"required,email"`
}
