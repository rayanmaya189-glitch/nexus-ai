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

func setupWorkflowHandler() *rest.WorkflowHandler {
	wfRepo := testutil.NewMockWorkflowRepository()
	stepRepo := testutil.NewMockWorkflowStepRepository()
	uc := usecases.NewWorkflowUseCase(wfRepo, stepRepo)
	return rest.NewWorkflowHandler(uc)
}

func TestWorkflowHandler_CreateWorkflow_Success(t *testing.T) {
	handler := setupWorkflowHandler()

	body := `{"tenant_id":1,"name":"ETL Pipeline","description":"Extract Transform Load","trigger_type":"scheduled"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateWorkflow(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "ETL Pipeline" {
		t.Errorf("expected name ETL Pipeline, got %v", data["name"])
	}
}

func TestWorkflowHandler_GetWorkflow_Success(t *testing.T) {
	handler := setupWorkflowHandler()

	createBody := `{"tenant_id":1,"name":"Pipeline"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/workflows", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateWorkflow(createRec, createReq)

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/workflows?id=1", nil)
	getRec := httptest.NewRecorder()
	handler.GetWorkflow(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", getRec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(getRec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "Pipeline" {
		t.Errorf("expected name Pipeline, got %v", data["name"])
	}
}

func TestWorkflowHandler_ListWorkflows_Success(t *testing.T) {
	handler := setupWorkflowHandler()

	body := `{"tenant_id":1,"name":"WF 1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.CreateWorkflow(rec, req)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/workflows?tenant_id=1", nil)
	listRec := httptest.NewRecorder()
	handler.ListWorkflows(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", listRec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(listRec.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("expected 1 workflow, got %d", len(data))
	}
}

func TestWorkflowHandler_UpdateWorkflow_Success(t *testing.T) {
	handler := setupWorkflowHandler()

	createBody := `{"tenant_id":1,"name":"Old Name"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/workflows", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateWorkflow(createRec, createReq)

	updateBody := `{"id":1,"name":"New Name"}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/workflows", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	handler.UpdateWorkflow(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", updateRec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(updateRec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "New Name" {
		t.Errorf("expected name New Name, got %v", data["name"])
	}
}

func TestWorkflowHandler_DeleteWorkflow_Success(t *testing.T) {
	handler := setupWorkflowHandler()

	createBody := `{"tenant_id":1,"name":"To Delete"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/workflows", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateWorkflow(createRec, createReq)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/workflows/1", nil)
	deleteRec := httptest.NewRecorder()
	handler.DeleteWorkflow(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", deleteRec.Code)
	}
}
