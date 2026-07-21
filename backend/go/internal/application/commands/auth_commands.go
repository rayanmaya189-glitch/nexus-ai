package commands

type LoginCommand struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterCommand struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	Name      string `json:"name" validate:"required"`
	TenantID  int64  `json:"tenant_id"`
}

type RefreshTokenCommand struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type CreateUserCommand struct {
	TenantID int64  `json:"tenant_id" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	RoleIDs  []int64 `json:"role_ids"`
}

type UpdateUserCommand struct {
	ID     int64  `json:"id" validate:"required"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type CreateTenantCommand struct {
	Name string `json:"name" validate:"required"`
	Slug string `json:"slug" validate:"required"`
	Plan string `json:"plan"`
}

type UpdateTenantCommand struct {
	ID       int64  `json:"id" validate:"required"`
	Name     string `json:"name"`
	Plan     string `json:"plan"`
	Status   string `json:"status"`
	Settings string `json:"settings"`
}

type CreateRoleCommand struct {
	TenantID    int64  `json:"tenant_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	PermissionIDs []int64 `json:"permission_ids"`
}

type AssignRoleCommand struct {
	UserID int64 `json:"user_id" validate:"required"`
	RoleID int64 `json:"role_id" validate:"required"`
}

type CreateAPIKeyCommand struct {
	TenantID    int64    `json:"tenant_id" validate:"required"`
	UserID      int64    `json:"user_id" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	Permissions []string `json:"permissions"`
	ExpiresAt   *string  `json:"expires_at"`
}
