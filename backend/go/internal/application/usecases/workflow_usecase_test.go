package usecases_test

import (
	"context"
	"testing"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/testutil"
)

func newTestWorkflowUseCase() (*usecases.WorkflowUseCase, *testutil.MockWorkflowRepository, *testutil.MockWorkflowStepRepository) {
	wfRepo := testutil.NewMockWorkflowRepository()
	stepRepo := testutil.NewMockWorkflowStepRepository()
	uc := usecases.NewWorkflowUseCase(wfRepo, stepRepo)
	return uc, wfRepo, stepRepo
}

func TestWorkflowUseCase_CreateWorkflow_Success(t *testing.T) {
	uc, _, _ := newTestWorkflowUseCase()

	cmd := commands.CreateWorkflowCommand{
		TenantID:    1,
		Name:        "Data Pipeline",
		Description: "Processes incoming data",
		TriggerType: "webhook",
	}
	wf, err := uc.CreateWorkflow(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if wf.Name != "Data Pipeline" {
		t.Errorf("expected name Data Pipeline, got %s", wf.Name)
	}
	if wf.Version != 1 {
		t.Errorf("expected version 1, got %d", wf.Version)
	}
	if wf.Status != "active" {
		t.Errorf("expected status active, got %s", wf.Status)
	}
}

func TestWorkflowUseCase_GetWorkflow_Success(t *testing.T) {
	uc, _, _ := newTestWorkflowUseCase()
	wf, _ := uc.CreateWorkflow(context.Background(), commands.CreateWorkflowCommand{
		TenantID: 1, Name: "Pipeline", TriggerType: "manual",
	})

	q := queries.GetWorkflowQuery{ID: wf.ID}
	found, err := uc.GetWorkflow(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Name != "Pipeline" {
		t.Errorf("expected name Pipeline, got %s", found.Name)
	}
}

func TestWorkflowUseCase_GetWorkflow_NotFound(t *testing.T) {
	uc, _, _ := newTestWorkflowUseCase()

	q := queries.GetWorkflowQuery{ID: 999}
	_, err := uc.GetWorkflow(context.Background(), q)
	if err == nil {
		t.Fatalf("expected error for not found workflow")
	}
}

func TestWorkflowUseCase_ListWorkflows_Success(t *testing.T) {
	uc, _, _ := newTestWorkflowUseCase()
	uc.CreateWorkflow(context.Background(), commands.CreateWorkflowCommand{
		TenantID: 1, Name: "WF 1",
	})
	uc.CreateWorkflow(context.Background(), commands.CreateWorkflowCommand{
		TenantID: 1, Name: "WF 2",
	})

	q := queries.ListWorkflowsQuery{TenantID: 1}
	wfs, err := uc.ListWorkflows(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(wfs) != 2 {
		t.Errorf("expected 2 workflows, got %d", len(wfs))
	}
}

func TestWorkflowUseCase_UpdateWorkflow_Success(t *testing.T) {
	uc, _, _ := newTestWorkflowUseCase()
	wf, _ := uc.CreateWorkflow(context.Background(), commands.CreateWorkflowCommand{
		TenantID: 1, Name: "Old Name",
	})

	cmd := commands.UpdateWorkflowCommand{
		ID:   wf.ID,
		Name: "New Name",
	}
	updated, err := uc.UpdateWorkflow(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name New Name, got %s", updated.Name)
	}
}

func TestWorkflowUseCase_DeleteWorkflow_Success(t *testing.T) {
	uc, wfRepo, _ := newTestWorkflowUseCase()
	wf, _ := uc.CreateWorkflow(context.Background(), commands.CreateWorkflowCommand{
		TenantID: 1, Name: "Pipeline",
	})

	err := uc.DeleteWorkflow(context.Background(), wf.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = wfRepo.FindByID(context.Background(), wf.ID)
	if err == nil {
		t.Errorf("expected error after deletion")
	}
}

func TestWorkflowUseCase_DeleteWorkflow_NotFound(t *testing.T) {
	uc, _, _ := newTestWorkflowUseCase()

	err := uc.DeleteWorkflow(context.Background(), 999)
	if err == nil {
		t.Fatalf("expected error for not found workflow")
	}
}
