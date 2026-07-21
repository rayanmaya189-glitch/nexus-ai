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

type AgentUseCase struct {
	agentRepo      repositories.AgentRepository
	executionRepo  repositories.AgentExecutionRepository
	stepRepo       repositories.AgentStepRepository
}

func NewAgentUseCase(
	agentRepo repositories.AgentRepository,
	executionRepo repositories.AgentExecutionRepository,
	stepRepo repositories.AgentStepRepository,
) *AgentUseCase {
	return &AgentUseCase{
		agentRepo:     agentRepo,
		executionRepo: executionRepo,
		stepRepo:      stepRepo,
	}
}

func (uc *AgentUseCase) GetAgent(ctx context.Context, q queries.GetAgentQuery) (*entities.Agent, error) {
	agent, err := uc.agentRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, nexuserrors.NotFound("Agent not found")
	}
	return agent, nil
}

func (uc *AgentUseCase) ListAgents(ctx context.Context, q queries.ListAgentsQuery) ([]*entities.Agent, error) {
	if q.AgentType != "" {
		return uc.agentRepo.FindByType(ctx, q.TenantID, q.AgentType)
	}
	return uc.agentRepo.FindByTenantID(ctx, q.TenantID)
}

func (uc *AgentUseCase) CreateAgent(ctx context.Context, cmd commands.CreateAgentCommand) (*entities.Agent, error) {
	agent := &entities.Agent{
		TenantID:     cmd.TenantID,
		Name:         cmd.Name,
		AgentType:    cmd.AgentType,
		Model:        cmd.Model,
		SystemPrompt: cmd.SystemPrompt,
		Capabilities: cmd.Capabilities,
		Status:       "active",
	}

	if err := uc.agentRepo.Create(ctx, agent); err != nil {
		return nil, nexuserrors.Internal("Failed to create agent")
	}
	return agent, nil
}

func (uc *AgentUseCase) UpdateAgent(ctx context.Context, cmd commands.UpdateAgentCommand) (*entities.Agent, error) {
	agent, err := uc.agentRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, nexuserrors.NotFound("Agent not found")
	}

	if cmd.Name != "" {
		agent.Name = cmd.Name
	}
	if cmd.AgentType != "" {
		agent.AgentType = cmd.AgentType
	}
	if cmd.Model != "" {
		agent.Model = cmd.Model
	}
	if cmd.SystemPrompt != "" {
		agent.SystemPrompt = cmd.SystemPrompt
	}
	if cmd.Capabilities != "" {
		agent.Capabilities = cmd.Capabilities
	}
	if cmd.Status != "" {
		agent.Status = cmd.Status
	}

	agent.UpdatedAt = time.Now()
	if err := uc.agentRepo.Update(ctx, agent); err != nil {
		return nil, nexuserrors.Internal("Failed to update agent")
	}
	return agent, nil
}

func (uc *AgentUseCase) DeleteAgent(ctx context.Context, id int64) error {
	_, err := uc.agentRepo.FindByID(ctx, id)
	if err != nil {
		return nexuserrors.NotFound("Agent not found")
	}
	return uc.agentRepo.Delete(ctx, id)
}

func (uc *AgentUseCase) GetExecution(ctx context.Context, q queries.GetAgentExecutionQuery) (*entities.AgentExecution, error) {
	exec, err := uc.executionRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, nexuserrors.NotFound("Execution not found")
	}
	return exec, nil
}

func (uc *AgentUseCase) ListExecutions(ctx context.Context, q queries.ListAgentExecutionsQuery) ([]*entities.AgentExecution, error) {
	return uc.executionRepo.FindByAgentID(ctx, q.AgentID)
}

func (uc *AgentUseCase) GetExecutionSteps(ctx context.Context, executionID int64) ([]*entities.AgentStep, error) {
	return uc.stepRepo.FindByExecutionID(ctx, executionID)
}
