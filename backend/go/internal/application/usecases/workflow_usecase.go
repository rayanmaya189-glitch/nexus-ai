package usecases

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/domain/repositories"
	nexuserrors "github.com/aeroxe/nexus-backend/pkg/errors"
)

type WorkflowUseCase struct {
	workflowRepo     repositories.WorkflowRepository
	workflowStepRepo repositories.WorkflowStepRepository
}

func NewWorkflowUseCase(
	workflowRepo repositories.WorkflowRepository,
	workflowStepRepo repositories.WorkflowStepRepository,
) *WorkflowUseCase {
	return &WorkflowUseCase{
		workflowRepo:     workflowRepo,
		workflowStepRepo: workflowStepRepo,
	}
}

func (uc *WorkflowUseCase) GetWorkflow(ctx context.Context, q queries.GetWorkflowQuery) (*entities.Workflow, error) {
	wf, err := uc.workflowRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, nexuserrors.NotFound("Workflow not found")
	}
	return wf, nil
}

func (uc *WorkflowUseCase) ListWorkflows(ctx context.Context, q queries.ListWorkflowsQuery) ([]*entities.Workflow, error) {
	return uc.workflowRepo.FindByTenantID(ctx, q.TenantID)
}

func (uc *WorkflowUseCase) CreateWorkflow(ctx context.Context, cmd commands.CreateWorkflowCommand) (*entities.Workflow, error) {
	wf := &entities.Workflow{
		TenantID:    cmd.TenantID,
		Name:        cmd.Name,
		Description: cmd.Description,
		TriggerType: cmd.TriggerType,
		Config:      cmd.Config,
		Status:      "active",
		Version:     1,
	}

	if wf.TriggerType == "" {
		wf.TriggerType = "manual"
	}

	if err := uc.workflowRepo.Create(ctx, wf); err != nil {
		return nil, nexuserrors.Internal("Failed to create workflow")
	}
	return wf, nil
}

func (uc *WorkflowUseCase) UpdateWorkflow(ctx context.Context, cmd commands.UpdateWorkflowCommand) (*entities.Workflow, error) {
	wf, err := uc.workflowRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, nexuserrors.NotFound("Workflow not found")
	}

	if cmd.Name != "" {
		wf.Name = cmd.Name
	}
	if cmd.Description != "" {
		wf.Description = cmd.Description
	}
	if cmd.TriggerType != "" {
		wf.TriggerType = cmd.TriggerType
	}
	if cmd.Config != "" {
		wf.Config = cmd.Config
	}
	if cmd.Status != "" {
		wf.Status = cmd.Status
	}

	wf.UpdatedAt = time.Now()
	if err := uc.workflowRepo.Update(ctx, wf); err != nil {
		return nil, nexuserrors.Internal("Failed to update workflow")
	}
	return wf, nil
}

func (uc *WorkflowUseCase) DeleteWorkflow(ctx context.Context, id int64) error {
	_, err := uc.workflowRepo.FindByID(ctx, id)
	if err != nil {
		return nexuserrors.NotFound("Workflow not found")
	}
	return uc.workflowRepo.Delete(ctx, id)
}

func (uc *WorkflowUseCase) GetWorkflowSteps(ctx context.Context, workflowID int64) ([]*entities.WorkflowStep, error) {
	return uc.workflowStepRepo.FindByWorkflowID(ctx, workflowID)
}
