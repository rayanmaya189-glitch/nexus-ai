package entities

import "time"

type PlatformConfiguration struct {
	ID           int64     `json:"id" db:"id"`
	ConfigKey    string    `json:"config_key" db:"config_key"`
	ConfigValue  string    `json:"config_value" db:"config_value"`
	Category     string    `json:"category" db:"category"`
	Description  string    `json:"description" db:"description"`
	DataType     string    `json:"data_type" db:"data_type"`
	IsSensitive  bool      `json:"is_sensitive" db:"is_sensitive"`
	DefaultValue string    `json:"default_value" db:"default_value"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type TenantConfiguration struct {
	ID          int64     `json:"id" db:"id"`
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	ConfigKey   string    `json:"config_key" db:"config_key"`
	ConfigValue string    `json:"config_value" db:"config_value"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
