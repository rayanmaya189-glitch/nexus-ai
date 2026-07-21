package usecases_test

import (
	"context"
	"testing"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/testutil"
)

func newTestAgentUseCase() (*usecases.AgentUseCase, *testutil.MockAgentRepository, *testutil.MockAgentExecutionRepository, *testutil.MockAgentStepRepository) {
	agentRepo := testutil.NewMockAgentRepository()
	execRepo := testutil.NewMockAgentExecutionRepository()
	stepRepo := testutil.NewMockAgentStepRepository()
	uc := usecases.NewAgentUseCase(agentRepo, execRepo, stepRepo)
	return uc, agentRepo, execRepo, stepRepo
}

func TestAgentUseCase_CreateAgent_Success(t *testing.T) {
	uc, _, _, _ := newTestAgentUseCase()

	cmd := commands.CreateAgentCommand{
		TenantID:  1,
		Name:      "Research Agent",
		AgentType: "research",
		Model:     "gpt-4",
	}
	agent, err := uc.CreateAgent(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if agent.Name != "Research Agent" {
		t.Errorf("expected name Research Agent, got %s", agent.Name)
	}
	if agent.Status != "active" {
		t.Errorf("expected status active, got %s", agent.Status)
	}
}

func TestAgentUseCase_GetAgent_Success(t *testing.T) {
	uc, agentRepo, _, _ := newTestAgentUseCase()
	cmd := commands.CreateAgentCommand{
		TenantID:  1,
		Name:      "Test Agent",
		AgentType: "chat",
		Model:     "gpt-4",
	}
	agent, _ := uc.CreateAgent(context.Background(), cmd)

	q := queries.GetAgentQuery{ID: agent.ID}
	found, err := uc.GetAgent(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Name != "Test Agent" {
		t.Errorf("expected name Test Agent, got %s", found.Name)
	}
	_ = agentRepo
}

func TestAgentUseCase_GetAgent_NotFound(t *testing.T) {
	uc, _, _, _ := newTestAgentUseCase()

	q := queries.GetAgentQuery{ID: 999}
	_, err := uc.GetAgent(context.Background(), q)
	if err == nil {
		t.Fatalf("expected error for not found agent")
	}
}

func TestAgentUseCase_ListAgents_Success(t *testing.T) {
	uc, _, _, _ := newTestAgentUseCase()
	uc.CreateAgent(context.Background(), commands.CreateAgentCommand{
		TenantID: 1, Name: "Agent 1", AgentType: "chat", Model: "gpt-4",
	})
	uc.CreateAgent(context.Background(), commands.CreateAgentCommand{
		TenantID: 1, Name: "Agent 2", AgentType: "research", Model: "gpt-4",
	})

	q := queries.ListAgentsQuery{TenantID: 1}
	agents, err := uc.ListAgents(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
}

func TestAgentUseCase_ListAgents_ByType(t *testing.T) {
	uc, _, _, _ := newTestAgentUseCase()
	uc.CreateAgent(context.Background(), commands.CreateAgentCommand{
		TenantID: 1, Name: "Chat Agent", AgentType: "chat", Model: "gpt-4",
	})
	uc.CreateAgent(context.Background(), commands.CreateAgentCommand{
		TenantID: 1, Name: "Research Agent", AgentType: "research", Model: "gpt-4",
	})

	q := queries.ListAgentsQuery{TenantID: 1, AgentType: "chat"}
	agents, err := uc.ListAgents(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}
	if agents[0].AgentType != "chat" {
		t.Errorf("expected agent type chat, got %s", agents[0].AgentType)
	}
}

func TestAgentUseCase_UpdateAgent_Success(t *testing.T) {
	uc, _, _, _ := newTestAgentUseCase()
	agent, _ := uc.CreateAgent(context.Background(), commands.CreateAgentCommand{
		TenantID: 1, Name: "Old Name", AgentType: "chat", Model: "gpt-4",
	})

	cmd := commands.UpdateAgentCommand{
		ID:   agent.ID,
		Name: "New Name",
	}
	updated, err := uc.UpdateAgent(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name New Name, got %s", updated.Name)
	}
}

func TestAgentUseCase_DeleteAgent_Success(t *testing.T) {
	uc, agentRepo, _, _ := newTestAgentUseCase()
	agent, _ := uc.CreateAgent(context.Background(), commands.CreateAgentCommand{
		TenantID: 1, Name: "Agent", AgentType: "chat", Model: "gpt-4",
	})

	err := uc.DeleteAgent(context.Background(), agent.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = agentRepo.FindByID(context.Background(), agent.ID)
	if err == nil {
		t.Errorf("expected error after deletion")
	}
}

func TestAgentUseCase_DeleteAgent_NotFound(t *testing.T) {
	uc, _, _, _ := newTestAgentUseCase()

	err := uc.DeleteAgent(context.Background(), 999)
	if err == nil {
		t.Fatalf("expected error for not found agent")
	}
}
