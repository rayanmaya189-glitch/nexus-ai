package repositories

import (
	"context"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
)

type PlatformConfigurationRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.PlatformConfiguration, error)
	FindByKey(ctx context.Context, key string) (*entities.PlatformConfiguration, error)
	FindByCategory(ctx context.Context, category string) ([]*entities.PlatformConfiguration, error)
	List(ctx context.Context) ([]*entities.PlatformConfiguration, error)
	Create(ctx context.Context, config *entities.PlatformConfiguration) error
	Update(ctx context.Context, config *entities.PlatformConfiguration) error
	Delete(ctx context.Context, id int64) error
}

type TenantConfigurationRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.TenantConfiguration, error)
	FindByTenantAndKey(ctx context.Context, tenantID int64, key string) (*entities.TenantConfiguration, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.TenantConfiguration, error)
	Create(ctx context.Context, config *entities.TenantConfiguration) error
	Update(ctx context.Context, config *entities.TenantConfiguration) error
	Delete(ctx context.Context, id int64) error
}
