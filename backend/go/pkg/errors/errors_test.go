package errors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	nexuserrors "github.com/aeroxe/nexus-backend/pkg/errors"
)

func TestAPIError_Error(t *testing.T) {
	err := nexuserrors.NewAPIError(http.StatusBadRequest, "TEST_ERROR", "Something went wrong")
	msg := err.Error()
	if !strings.Contains(msg, "TEST_ERROR") {
		t.Errorf("expected error message to contain TEST_ERROR, got %s", msg)
	}
	if !strings.Contains(msg, "Something went wrong") {
		t.Errorf("expected error message to contain 'Something went wrong', got %s", msg)
	}
}

func TestAPIError_ToJSON(t *testing.T) {
	err := nexuserrors.NewAPIError(http.StatusBadRequest, "TEST_ERROR", "Bad input")
	j := err.ToJSON()

	errObj, ok := j["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error object in JSON")
	}
	if errObj["code"] != "TEST_ERROR" {
		t.Errorf("expected code TEST_ERROR, got %v", errObj["code"])
	}
	if errObj["message"] != "Bad input" {
		t.Errorf("expected message 'Bad input', got %v", errObj["message"])
	}
}

func TestUnauthorized(t *testing.T) {
	err := nexuserrors.Unauthorized("Not logged in")
	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", err.StatusCode)
	}
	if err.Code != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %s", err.Code)
	}
}

func TestForbidden(t *testing.T) {
	err := nexuserrors.Forbidden("No access")
	if err.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", err.StatusCode)
	}
	if err.Code != "FORBIDDEN" {
		t.Errorf("expected code FORBIDDEN, got %s", err.Code)
	}
}

func TestNotFound(t *testing.T) {
	err := nexuserrors.NotFound("Resource missing")
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", err.StatusCode)
	}
	if err.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", err.Code)
	}
}

func TestBadRequest(t *testing.T) {
	err := nexuserrors.BadRequest("Invalid input")
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.StatusCode)
	}
	if err.Code != "BAD_REQUEST" {
		t.Errorf("expected code BAD_REQUEST, got %s", err.Code)
	}
}

func TestConflict(t *testing.T) {
	err := nexuserrors.Conflict("Already exists")
	if err.StatusCode != http.StatusConflict {
		t.Errorf("expected status 409, got %d", err.StatusCode)
	}
	if err.Code != "CONFLICT" {
		t.Errorf("expected code CONFLICT, got %s", err.Code)
	}
}

func TestInternal(t *testing.T) {
	err := nexuserrors.Internal("Server error")
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", err.StatusCode)
	}
	if err.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got %s", err.Code)
	}
}

func TestAPIError_WriteJSON(t *testing.T) {
	err := nexuserrors.Unauthorized("No token")
	rec := httptest.NewRecorder()

	err.WriteJSON(rec)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED in body, got %v", errObj["code"])
	}
}
