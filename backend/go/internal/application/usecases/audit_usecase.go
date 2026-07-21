package usecases

import (
	"context"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/domain/repositories"
	nexuserrors "github.com/aeroxe/nexus-backend/pkg/errors"
)

type AuditUseCase struct {
	auditRepo repositories.AuditLogRepository
}

func NewAuditUseCase(auditRepo repositories.AuditLogRepository) *AuditUseCase {
	return &AuditUseCase{auditRepo: auditRepo}
}

func (uc *AuditUseCase) CreateAuditLog(ctx context.Context, cmd commands.CreateAuditLogCommand) (*entities.AuditLog, error) {
	log := &entities.AuditLog{
		TenantID:     cmd.TenantID,
		UserID:       cmd.UserID,
		Action:       cmd.Action,
		ResourceType: cmd.ResourceType,
		ResourceID:   cmd.ResourceID,
		Details:      cmd.Details,
		IPAddress:    cmd.IPAddress,
		UserAgent:    cmd.UserAgent,
		Status:       cmd.Status,
	}

	if log.Status == "" {
		log.Status = "success"
	}

	if err := uc.auditRepo.Create(ctx, log); err != nil {
		return nil, nexuserrors.Internal("Failed to create audit log")
	}
	return log, nil
}

func (uc *AuditUseCase) GetAuditLog(ctx context.Context, q queries.GetAuditLogQuery) (*entities.AuditLog, error) {
	log, err := uc.auditRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, nexuserrors.NotFound("Audit log not found")
	}
	return log, nil
}

func (uc *AuditUseCase) ListAuditLogs(ctx context.Context, q queries.ListAuditLogsQuery) ([]*entities.AuditLog, int64, error) {
	page := q.Page
	if page <= 0 {
		page = 1
	}
	perPage := q.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	logs, count, err := uc.auditRepo.FindByTenantID(ctx, q.TenantID, perPage, offset)
	if err != nil {
		return nil, 0, nexuserrors.Internal("Failed to list audit logs")
	}

	return logs, count, nil
}
