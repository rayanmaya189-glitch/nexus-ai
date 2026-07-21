package repositories

import (
	"context"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
)

type IntegrationRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Integration, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Integration, error)
	FindByType(ctx context.Context, tenantID int64, integrationType string) ([]*entities.Integration, error)
	Create(ctx context.Context, integration *entities.Integration) error
	Update(ctx context.Context, integration *entities.Integration) error
	Delete(ctx context.Context, id int64) error
}

type MCPToolRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.MCPTool, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.MCPTool, error)
	FindByIntegrationID(ctx context.Context, integrationID int64) ([]*entities.MCPTool, error)
	GetEnabledTools(ctx context.Context, tenantID int64) ([]*entities.MCPTool, error)
	Create(ctx context.Context, tool *entities.MCPTool) error
	Update(ctx context.Context, tool *entities.MCPTool) error
	Delete(ctx context.Context, id int64) error
}

type MCPToolInvocationRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.MCPToolInvocation, error)
	FindByToolID(ctx context.Context, toolID int64) ([]*entities.MCPToolInvocation, error)
	Create(ctx context.Context, invocation *entities.MCPToolInvocation) error
}
