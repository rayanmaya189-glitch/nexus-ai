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

func setupDocumentHandler() *rest.DocumentHandler {
	docRepo := testutil.NewMockDocumentRepository()
	chunkRepo := testutil.NewMockDocumentChunkRepository()
	docSetRepo := testutil.NewMockDocumentSetRepository()
	uc := usecases.NewDocumentUseCase(docRepo, chunkRepo, docSetRepo)
	return rest.NewDocumentHandler(uc)
}

func TestDocumentHandler_CreateDocument_Success(t *testing.T) {
	handler := setupDocumentHandler()

	body := `{"tenant_id":1,"title":"API Docs","content":"Endpoint documentation","doc_type":"markdown"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rag/documents", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateDocument(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["title"] != "API Docs" {
		t.Errorf("expected title API Docs, got %v", data["title"])
	}
}

func TestDocumentHandler_GetDocument_Success(t *testing.T) {
	handler := setupDocumentHandler()

	createBody := `{"tenant_id":1,"title":"Test Doc","content":"Content here"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/rag/documents", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateDocument(createRec, createReq)

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/rag/documents?id=1", nil)
	getRec := httptest.NewRecorder()
	handler.GetDocument(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", getRec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(getRec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["title"] != "Test Doc" {
		t.Errorf("expected title Test Doc, got %v", data["title"])
	}
}

func TestDocumentHandler_ListDocuments_Success(t *testing.T) {
	handler := setupDocumentHandler()

	body := `{"tenant_id":1,"title":"Doc 1","content":"Content 1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rag/documents", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.CreateDocument(rec, req)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/rag/documents?tenant_id=1", nil)
	listRec := httptest.NewRecorder()
	handler.ListDocuments(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", listRec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(listRec.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("expected 1 document, got %d", len(data))
	}
}

func TestDocumentHandler_UpdateDocument_Success(t *testing.T) {
	handler := setupDocumentHandler()

	createBody := `{"tenant_id":1,"title":"Old Title","content":"Content"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/rag/documents", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateDocument(createRec, createReq)

	updateBody := `{"id":1,"title":"Updated Title"}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/rag/documents", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	handler.UpdateDocument(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", updateRec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(updateRec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["title"] != "Updated Title" {
		t.Errorf("expected title Updated Title, got %v", data["title"])
	}
}

func TestDocumentHandler_DeleteDocument_Success(t *testing.T) {
	handler := setupDocumentHandler()

	createBody := `{"tenant_id":1,"title":"To Delete","content":"Content"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/rag/documents", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.CreateDocument(createRec, createReq)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/rag/documents/1", nil)
	deleteRec := httptest.NewRecorder()
	handler.DeleteDocument(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", deleteRec.Code)
	}
}
