package entities

import "time"

type User struct {
	ID          int64     `json:"id" db:"id"`
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	Email       string    `json:"email" db:"email"`
	PasswordHash string  `json:"-" db:"password_hash"`
	Name        string    `json:"name" db:"name"`
	Status      string    `json:"status" db:"status"`
	Avatar      string    `json:"avatar" db:"avatar"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Role struct {
	ID          int64     `json:"id" db:"id"`
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Permission struct {
	ID          int64     `json:"id" db:"id"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
}

type Tenant struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Plan        string    `json:"plan" db:"plan"`
	Status      string    `json:"status" db:"status"`
	Settings    string    `json:"settings" db:"settings"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type APIKey struct {
	ID          int64     `json:"id" db:"id"`
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	KeyHash     string    `json:"-" db:"key_hash"`
	Permissions string    `json:"permissions" db:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
