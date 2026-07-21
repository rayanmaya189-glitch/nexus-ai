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

func setupAgentHandler() *rest.AgentHandler {
	agentRepo := testutil.NewMockAgentRepository()
	execRepo := testutil.NewMockAgentExecutionRepository()
	stepRepo := testutil.NewMockAgentStepRepository()
	uc := usecases.NewAgentUseCase(agentRepo, execRepo, stepRepo)
	return rest.NewAgentHandler(uc)
}

func TestAgentHandler_CreateAgent_Success(t *testing.T) {
	handler := setupAgentHandler()

	body := `{"tenant_id":1,"name":"Test Agent","agent_type":"chat","model":"gpt-4"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateAgent(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "Test Agent" {
		t.Errorf("expected name Test Agent, got %v", data["name"])
	}
}

func TestAgentHandler_GetAgent_Success(t *testing.T) {
	handler := setupAgentHandler()

	createBody := `{"tenant_id":1,"name":"My Agent","agent_type":"research","model":"gpt-4"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/agents", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateAgent(createRec, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRec.Body.Bytes(), &createResp)
	agentData := createResp["data"].(map[string]interface{})
	id := int(agentData["id"].(float64))

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/agents?id=1", nil)
	getRec := httptest.NewRecorder()
	handler.GetAgent(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", getRec.Code)
	}

	var getResp map[string]interface{}
	json.Unmarshal(getRec.Body.Bytes(), &getResp)
	data := getResp["data"].(map[string]interface{})
	if data["name"] != "My Agent" {
		t.Errorf("expected name My Agent, got %v", data["name"])
	}
	_ = id
}

func TestAgentHandler_ListAgents_Success(t *testing.T) {
	handler := setupAgentHandler()

	body := `{"tenant_id":1,"name":"Agent 1","agent_type":"chat","model":"gpt-4"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.CreateAgent(rec, req)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/agents?tenant_id=1", nil)
	listRec := httptest.NewRecorder()
	handler.ListAgents(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", listRec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(listRec.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("expected 1 agent, got %d", len(data))
	}
}

func TestAgentHandler_UpdateAgent_Success(t *testing.T) {
	handler := setupAgentHandler()

	createBody := `{"tenant_id":1,"name":"Old Name","agent_type":"chat","model":"gpt-4"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/agents", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateAgent(createRec, createReq)

	updateBody := `{"id":1,"name":"New Name"}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/agents", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	handler.UpdateAgent(updateRec, updateReq)

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

func TestAgentHandler_DeleteAgent_Success(t *testing.T) {
	handler := setupAgentHandler()

	createBody := `{"tenant_id":1,"name":"To Delete","agent_type":"chat","model":"gpt-4"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/agents", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateAgent(createRec, createReq)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/1", nil)
	deleteRec := httptest.NewRecorder()
	handler.DeleteAgent(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", deleteRec.Code)
	}
}
