package repositories

import (
	"context"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
)

type AgentRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Agent, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Agent, error)
	FindByType(ctx context.Context, tenantID int64, agentType string) ([]*entities.Agent, error)
	Create(ctx context.Context, agent *entities.Agent) error
	Update(ctx context.Context, agent *entities.Agent) error
	Delete(ctx context.Context, id int64) error
}

type AgentExecutionRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.AgentExecution, error)
	FindByAgentID(ctx context.Context, agentID int64) ([]*entities.AgentExecution, error)
	Create(ctx context.Context, execution *entities.AgentExecution) error
	Update(ctx context.Context, execution *entities.AgentExecution) error
}

type AgentStepRepository interface {
	FindByExecutionID(ctx context.Context, executionID int64) ([]*entities.AgentStep, error)
	Create(ctx context.Context, step *entities.AgentStep) error
	Update(ctx context.Context, step *entities.AgentStep) error
}
