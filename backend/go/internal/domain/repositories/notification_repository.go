package repositories

import (
	"context"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
)

type NotificationRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Notification, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Notification, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error)
	Create(ctx context.Context, notification *entities.Notification) error
	Update(ctx context.Context, notification *entities.Notification) error
	Delete(ctx context.Context, id int64) error
	MarkAsRead(ctx context.Context, id int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
}

type NotificationPreferenceRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.NotificationPreference, error)
	FindByTenantAndUser(ctx context.Context, tenantID, userID int64) (*entities.NotificationPreference, error)
	Create(ctx context.Context, pref *entities.NotificationPreference) error
	Update(ctx context.Context, pref *entities.NotificationPreference) error
	Delete(ctx context.Context, id int64) error
}

type NotificationTemplateRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.NotificationTemplate, error)
	FindByName(ctx context.Context, name string) (*entities.NotificationTemplate, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.NotificationTemplate, error)
	Create(ctx context.Context, tmpl *entities.NotificationTemplate) error
	Update(ctx context.Context, tmpl *entities.NotificationTemplate) error
	Delete(ctx context.Context, id int64) error
}
