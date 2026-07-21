package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/interfaces/rest"
	"github.com/aeroxe/nexus-backend/internal/testutil"
	"github.com/aeroxe/nexus-backend/pkg/auth"
)

func setupAuthHandler() (*rest.AuthHandler, *testutil.MockUserRepository, *testutil.MockTenantRepository) {
	userRepo := testutil.NewMockUserRepository()
	tenantRepo := testutil.NewMockTenantRepository()
	roleRepo := testutil.NewMockRoleRepository()
	jwtManager := auth.NewJWTManager("test-secret", "nexus-test", 1*time.Hour, 7*24*time.Hour)
	uc := usecases.NewAuthUseCase(userRepo, tenantRepo, roleRepo, jwtManager)
	handler := rest.NewAuthHandler(uc)
	return handler, userRepo, tenantRepo
}

func seedHandlerUser(t *testing.T, repo *testutil.MockUserRepository, email, password string, tenantID int64) *entities.User {
	t.Helper()
	hash, _ := auth.HashPassword(password)
	user := &entities.User{
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: hash,
		Name:         "Test User",
		Status:       "active",
	}
	repo.Create(context.Background(), user)
	return user
}

func TestAuthHandler_Login_Success(t *testing.T) {
	handler, userRepo, _ := setupAuthHandler()
	hash, _ := auth.HashPassword("password123")
	user := &entities.User{
		TenantID:     1,
		Email:        "test@example.com",
		PasswordHash: hash,
		Name:         "Test User",
		Status:       "active",
	}
	userRepo.Create(context.Background(), user)

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data in response")
	}
	if data["access_token"] == nil || data["access_token"] == "" {
		t.Errorf("expected non-empty access_token")
	}
	_ = seedHandlerUser
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	handler, _, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_Login_Unauthorized(t *testing.T) {
	handler, _, _ := setupAuthHandler()

	body := `{"email":"nobody@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	handler, _, _ := setupAuthHandler()

	body := `{"email":"new@example.com","password":"password123","name":"New User","tenant_id":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["email"] != "new@example.com" {
		t.Errorf("expected email new@example.com, got %v", data["email"])
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	handler, userRepo, _ := setupAuthHandler()
	seedHandlerUser(t, userRepo, "existing@example.com", "password123", 1)

	body := `{"email":"existing@example.com","password":"password123","name":"Dup User","tenant_id":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestAuthHandler_ListUsers_Success(t *testing.T) {
	handler, userRepo, _ := setupAuthHandler()
	seedHandlerUser(t, userRepo, "u1@example.com", "password123", 1)
	seedHandlerUser(t, userRepo, "u2@example.com", "password123", 1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?tenant_id=1", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["data"] == nil {
		t.Errorf("expected data in response")
	}
	if resp["meta"] == nil {
		t.Errorf("expected meta in response")
	}
}

func TestAuthHandler_GetUser_Success(t *testing.T) {
	handler, userRepo, _ := setupAuthHandler()
	user := seedHandlerUser(t, userRepo, "test@example.com", "password123", 1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?id=1", nil)
	rec := httptest.NewRecorder()

	handler.GetUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["email"] != "test@example.com" {
		t.Errorf("expected email test@example.com, got %v", data["email"])
	}
	_ = user
}

func TestAuthHandler_CreateUser_Success(t *testing.T) {
	handler, _, _ := setupAuthHandler()

	body := `{"tenant_id":1,"email":"admin@example.com","name":"Admin","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestAuthHandler_CreateTenant_Success(t *testing.T) {
	handler, _, _ := setupAuthHandler()

	body := `{"name":"Acme Corp","slug":"acme","plan":"pro"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateTenant(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "Acme Corp" {
		t.Errorf("expected name Acme Corp, got %v", data["name"])
	}
}

func TestAuthHandler_ListTenants_Success(t *testing.T) {
	handler, _, tenantRepo := setupAuthHandler()
	tenantRepo.Create(context.Background(), &entities.Tenant{Name: "Org 1", Slug: "org1", Plan: "free", Status: "active"})
	tenantRepo.Create(context.Background(), &entities.Tenant{Name: "Org 2", Slug: "org2", Plan: "pro", Status: "active"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants?page=1&per_page=10", nil)
	rec := httptest.NewRecorder()

	handler.ListTenants(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["meta"] == nil {
		t.Errorf("expected meta in response")
	}
}
