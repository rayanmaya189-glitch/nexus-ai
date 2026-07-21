package entities

import "time"

type Notification struct {
	ID        int64     `json:"id" db:"id"`
	TenantID  int64     `json:"tenant_id" db:"tenant_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message" db:"message"`
	Channel   string    `json:"channel" db:"channel"`
	Status    string    `json:"status" db:"status"`
	Metadata  string    `json:"metadata" db:"metadata"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type NotificationPreference struct {
	ID              int64     `json:"id" db:"id"`
	TenantID        int64     `json:"tenant_id" db:"tenant_id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	EmailEnabled    bool      `json:"email_enabled" db:"email_enabled"`
	PushEnabled     bool      `json:"push_enabled" db:"push_enabled"`
	WebhookEnabled  bool      `json:"webhook_enabled" db:"webhook_enabled"`
	WebhookURL      string    `json:"webhook_url" db:"webhook_url"`
	EmailAddress    string    `json:"email_address" db:"email_address"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type NotificationTemplate struct {
	ID              int64     `json:"id" db:"id"`
	TenantID        *int64    `json:"tenant_id" db:"tenant_id"`
	Name            string    `json:"name" db:"name"`
	Channel         string    `json:"channel" db:"channel"`
	SubjectTemplate string    `json:"subject_template" db:"subject_template"`
	BodyTemplate    string    `json:"body_template" db:"body_template"`
	IsSystem        bool      `json:"is_system" db:"is_system"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
