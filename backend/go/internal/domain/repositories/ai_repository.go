package repositories

import (
	"context"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
)

type AISessionRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.AISession, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.AISession, error)
	Create(ctx context.Context, session *entities.AISession) error
	Update(ctx context.Context, session *entities.AISession) error
}

type AIRequestRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.AIRequest, error)
	FindBySessionID(ctx context.Context, sessionID int64) ([]*entities.AIRequest, error)
	Create(ctx context.Context, request *entities.AIRequest) error
}

type MemoryRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Memory, error)
	FindByAgentID(ctx context.Context, agentID int64) ([]*entities.Memory, error)
	FindByTenantAndUser(ctx context.Context, tenantID, userID int64) ([]*entities.Memory, error)
	Create(ctx context.Context, memory *entities.Memory) error
	Update(ctx context.Context, memory *entities.Memory) error
	Delete(ctx context.Context, id int64) error
}

type WorkflowRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Workflow, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Workflow, error)
	Create(ctx context.Context, workflow *entities.Workflow) error
	Update(ctx context.Context, workflow *entities.Workflow) error
	Delete(ctx context.Context, id int64) error
}

type WorkflowStepRepository interface {
	FindByWorkflowID(ctx context.Context, workflowID int64) ([]*entities.WorkflowStep, error)
	Create(ctx context.Context, step *entities.WorkflowStep) error
	Update(ctx context.Context, step *entities.WorkflowStep) error
}

type AIModelRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.AIModel, error)
	FindByCategory(ctx context.Context, category string) ([]*entities.AIModel, error)
	List(ctx context.Context) ([]*entities.AIModel, error)
	Create(ctx context.Context, model *entities.AIModel) error
	Update(ctx context.Context, model *entities.AIModel) error
}

type AuditLogRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.AuditLog, error)
	FindByTenantID(ctx context.Context, tenantID int64, limit, offset int) ([]*entities.AuditLog, int64, error)
	Create(ctx context.Context, log *entities.AuditLog) error
	Count(ctx context.Context, tenantID int64) (int64, error)
}
