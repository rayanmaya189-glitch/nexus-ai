package httputil_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aeroxe/nexus-backend/pkg/httputil"
)

func TestWriteSuccess(t *testing.T) {
	rec := httptest.NewRecorder()

	httputil.WriteSuccess(rec, map[string]string{"name": "test"})

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp httputil.Response
	json.Unmarshal(rec.Body.Bytes(), &resp)

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data in response")
	}
	if data["name"] != "test" {
		t.Errorf("expected name test, got %v", data["name"])
	}
}

func TestWriteCreated(t *testing.T) {
	rec := httptest.NewRecorder()

	httputil.WriteCreated(rec, map[string]int64{"id": 1})

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp httputil.Response
	json.Unmarshal(rec.Body.Bytes(), &resp)

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data in response")
	}
	if data["id"] != float64(1) {
		t.Errorf("expected id 1, got %v", data["id"])
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()

	httputil.WriteError(rec, http.StatusBadRequest, "INVALID_INPUT", "Bad data")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp httputil.Response
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Error == nil {
		t.Fatalf("expected error in response")
	}
	if resp.Error.Code != "INVALID_INPUT" {
		t.Errorf("expected code INVALID_INPUT, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "Bad data" {
		t.Errorf("expected message 'Bad data', got %s", resp.Error.Message)
	}
}

func TestWritePaginated(t *testing.T) {
	rec := httptest.NewRecorder()

	items := []map[string]string{{"name": "a"}, {"name": "b"}}
	httputil.WritePaginated(rec, items, 1, 10, 25)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp httputil.Response
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Data == nil {
		t.Errorf("expected data in response")
	}
	if resp.Meta == nil {
		t.Fatalf("expected meta in response")
	}
	if resp.Meta.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Meta.Page)
	}
	if resp.Meta.PerPage != 10 {
		t.Errorf("expected per_page 10, got %d", resp.Meta.PerPage)
	}
	if resp.Meta.Total != 25 {
		t.Errorf("expected total 25, got %d", resp.Meta.Total)
	}
	if resp.Meta.TotalPages != 3 {
		t.Errorf("expected total_pages 3, got %d", resp.Meta.TotalPages)
	}
}
