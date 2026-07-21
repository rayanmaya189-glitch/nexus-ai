package usecases_test

import (
	"context"
	"testing"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/testutil"
)

func newTestAuditUseCase() (*usecases.AuditUseCase, *testutil.MockAuditLogRepository) {
	auditRepo := testutil.NewMockAuditLogRepository()
	uc := usecases.NewAuditUseCase(auditRepo)
	return uc, auditRepo
}

func TestAuditUseCase_CreateAuditLog_Success(t *testing.T) {
	uc, _ := newTestAuditUseCase()

	cmd := commands.CreateAuditLogCommand{
		TenantID:     1,
		UserID:       10,
		Action:       "create",
		ResourceType: "agent",
		ResourceID:   "42",
		Details:      "Created new agent",
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}
	log, err := uc.CreateAuditLog(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if log.Action != "create" {
		t.Errorf("expected action create, got %s", log.Action)
	}
	if log.Status != "success" {
		t.Errorf("expected status success, got %s", log.Status)
	}
}

func TestAuditUseCase_GetAuditLog_Success(t *testing.T) {
	uc, _ := newTestAuditUseCase()
	log, _ := uc.CreateAuditLog(context.Background(), commands.CreateAuditLogCommand{
		TenantID: 1, UserID: 1, Action: "login", ResourceType: "user",
	})

	q := queries.GetAuditLogQuery{ID: log.ID}
	found, err := uc.GetAuditLog(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Action != "login" {
		t.Errorf("expected action login, got %s", found.Action)
	}
}

func TestAuditUseCase_GetAuditLog_NotFound(t *testing.T) {
	uc, _ := newTestAuditUseCase()

	q := queries.GetAuditLogQuery{ID: 999}
	_, err := uc.GetAuditLog(context.Background(), q)
	if err == nil {
		t.Fatalf("expected error for not found audit log")
	}
}

func TestAuditUseCase_ListAuditLogs_Success(t *testing.T) {
	uc, _ := newTestAuditUseCase()
	for i := 0; i < 5; i++ {
		uc.CreateAuditLog(context.Background(), commands.CreateAuditLogCommand{
			TenantID: 1, UserID: 1, Action: "action", ResourceType: "resource",
		})
	}

	q := queries.ListAuditLogsQuery{TenantID: 1, Page: 1, PerPage: 10}
	logs, total, err := uc.ListAuditLogs(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(logs) != 5 {
		t.Errorf("expected 5 logs, got %d", len(logs))
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
}

func TestAuditUseCase_ListAuditLogs_Pagination(t *testing.T) {
	uc, _ := newTestAuditUseCase()
	for i := 0; i < 10; i++ {
		uc.CreateAuditLog(context.Background(), commands.CreateAuditLogCommand{
			TenantID: 1, UserID: 1, Action: "action", ResourceType: "resource",
		})
	}

	q := queries.ListAuditLogsQuery{TenantID: 1, Page: 2, PerPage: 3}
	logs, total, err := uc.ListAuditLogs(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(logs) != 3 {
		t.Errorf("expected 3 logs on page 2, got %d", len(logs))
	}
	if total != 10 {
		t.Errorf("expected total 10, got %d", total)
	}
}
