package repositories

import (
	"context"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
)

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.User, error)
	Create(ctx context.Context, user *entities.User) error
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context, tenantID int64) (int64, error)
}

type TenantRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*entities.Tenant, error)
	Create(ctx context.Context, tenant *entities.Tenant) error
	Update(ctx context.Context, tenant *entities.Tenant) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, perPage int) ([]*entities.Tenant, int64, error)
}

type RoleRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Role, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Role, error)
	Create(ctx context.Context, role *entities.Role) error
	Update(ctx context.Context, role *entities.Role) error
	Delete(ctx context.Context, id int64) error
	AssignRole(ctx context.Context, userID, roleID int64) error
	RemoveRole(ctx context.Context, userID, roleID int64) error
}

type APIKeyRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.APIKey, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.APIKey, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.APIKey, error)
	Create(ctx context.Context, apiKey *entities.APIKey) error
	Update(ctx context.Context, apiKey *entities.APIKey) error
	Delete(ctx context.Context, id int64) error
}
