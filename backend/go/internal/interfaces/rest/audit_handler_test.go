package rest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/interfaces/rest"
	"github.com/aeroxe/nexus-backend/internal/testutil"
)

func setupAuditHandler() *rest.AuditHandler {
	auditRepo := testutil.NewMockAuditLogRepository()
	uc := usecases.NewAuditUseCase(auditRepo)
	return rest.NewAuditHandler(uc)
}

func TestAuditHandler_CreateAuditLog_Success(t *testing.T) {
	handler := setupAuditHandler()

	body := `{"tenant_id":1,"user_id":10,"action":"create","resource_type":"agent","resource_id":"42","details":"Created agent","ip_address":"127.0.0.1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/audit-logs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateAuditLog(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["action"] != "create" {
		t.Errorf("expected action create, got %v", data["action"])
	}
}

func TestAuditHandler_GetAuditLog_Success(t *testing.T) {
	handler := setupAuditHandler()

	createBody := `{"tenant_id":1,"user_id":1,"action":"login","resource_type":"user"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/audit-logs", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateAuditLog(createRec, createReq)

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?id=1", nil)
	getRec := httptest.NewRecorder()
	handler.GetAuditLog(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", getRec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(getRec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["action"] != "login" {
		t.Errorf("expected action login, got %v", data["action"])
	}
}

func TestAuditHandler_ListAuditLogs_Success(t *testing.T) {
	handler := setupAuditHandler()

	body := `{"tenant_id":1,"user_id":1,"action":"view","resource_type":"document"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/audit-logs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.CreateAuditLog(rec, req)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?tenant_id=1&page=1&per_page=10", nil)
	listRec := httptest.NewRecorder()
	handler.ListAuditLogs(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", listRec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(listRec.Body.Bytes(), &resp)
	if resp["meta"] == nil {
		t.Errorf("expected meta in response")
	}
}
