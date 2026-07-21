package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/testutil"
	"github.com/aeroxe/nexus-backend/pkg/auth"
)

func newTestAuthUseCase() (*usecases.AuthUseCase, *testutil.MockUserRepository, *testutil.MockTenantRepository, *testutil.MockRoleRepository) {
	userRepo := testutil.NewMockUserRepository()
	tenantRepo := testutil.NewMockTenantRepository()
	roleRepo := testutil.NewMockRoleRepository()
	jwtManager := auth.NewJWTManager("test-secret-key-for-testing", "nexus-test", 1*time.Hour, 7*24*time.Hour)
	uc := usecases.NewAuthUseCase(userRepo, tenantRepo, roleRepo, jwtManager)
	return uc, userRepo, tenantRepo, roleRepo
}

func seedTestUser(t *testing.T, repo *testutil.MockUserRepository, email, password string, tenantID int64) *entities.User {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := &entities.User{
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: hash,
		Name:         "Test User",
		Status:       "active",
	}
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return user
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	uc, userRepo, _, _ := newTestAuthUseCase()
	seedTestUser(t, userRepo, "test@example.com", "password123", 1)

	cmd := commands.LoginCommand{Email: "test@example.com", Password: "password123"}
	result, err := uc.Login(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.AccessToken == "" {
		t.Errorf("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Errorf("expected non-empty refresh token")
	}
}

func TestAuthUseCase_Login_InvalidCredentials(t *testing.T) {
	uc, userRepo, _, _ := newTestAuthUseCase()
	seedTestUser(t, userRepo, "test@example.com", "password123", 1)

	cmd := commands.LoginCommand{Email: "test@example.com", Password: "wrongpassword"}
	_, err := uc.Login(context.Background(), cmd)
	if err == nil {
		t.Fatalf("expected error for invalid credentials")
	}
}

func TestAuthUseCase_Login_InactiveUser(t *testing.T) {
	uc, userRepo, _, _ := newTestAuthUseCase()
	hash, _ := auth.HashPassword("password123")
	user := &entities.User{
		TenantID:     1,
		Email:        "inactive@example.com",
		PasswordHash: hash,
		Name:         "Inactive User",
		Status:       "inactive",
	}
	userRepo.Create(context.Background(), user)

	cmd := commands.LoginCommand{Email: "inactive@example.com", Password: "password123"}
	_, err := uc.Login(context.Background(), cmd)
	if err == nil {
		t.Fatalf("expected error for inactive user")
	}
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	uc, _, _, _ := newTestAuthUseCase()

	cmd := commands.RegisterCommand{
		Email:    "new@example.com",
		Password: "password123",
		Name:     "New User",
		TenantID: 1,
	}
	user, err := uc.Register(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %s", user.Email)
	}
	if user.Status != "active" {
		t.Errorf("expected status active, got %s", user.Status)
	}
}

func TestAuthUseCase_Register_DuplicateEmail(t *testing.T) {
	uc, userRepo, _, _ := newTestAuthUseCase()
	seedTestUser(t, userRepo, "existing@example.com", "password123", 1)

	cmd := commands.RegisterCommand{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Dup User",
		TenantID: 1,
	}
	_, err := uc.Register(context.Background(), cmd)
	if err == nil {
		t.Fatalf("expected error for duplicate email")
	}
}

func TestAuthUseCase_CreateUser_Success(t *testing.T) {
	uc, _, _, _ := newTestAuthUseCase()

	cmd := commands.CreateUserCommand{
		TenantID: 1,
		Email:    "admin@example.com",
		Name:     "Admin User",
		Password: "password123",
	}
	user, err := uc.CreateUser(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "admin@example.com" {
		t.Errorf("expected email admin@example.com, got %s", user.Email)
	}
}

func TestAuthUseCase_UpdateUser_Success(t *testing.T) {
	uc, userRepo, _, _ := newTestAuthUseCase()
	user := seedTestUser(t, userRepo, "test@example.com", "password123", 1)

	cmd := commands.UpdateUserCommand{
		ID:   user.ID,
		Name: "Updated Name",
	}
	updated, err := uc.UpdateUser(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected name Updated Name, got %s", updated.Name)
	}
}

func TestAuthUseCase_DeleteUser_Success(t *testing.T) {
	uc, userRepo, _, _ := newTestAuthUseCase()
	user := seedTestUser(t, userRepo, "test@example.com", "password123", 1)

	err := uc.DeleteUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = userRepo.FindByID(context.Background(), user.ID)
	if err == nil {
		t.Errorf("expected error after deletion")
	}
}

func TestAuthUseCase_DeleteUser_NotFound(t *testing.T) {
	uc, _, _, _ := newTestAuthUseCase()

	err := uc.DeleteUser(context.Background(), 999)
	if err == nil {
		t.Fatalf("expected error for not found user")
	}
}

func TestAuthUseCase_GetUser_Success(t *testing.T) {
	uc, userRepo, _, _ := newTestAuthUseCase()
	user := seedTestUser(t, userRepo, "test@example.com", "password123", 1)

	q := queries.GetUserQuery{ID: user.ID}
	found, err := uc.GetUser(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", found.Email)
	}
}

func TestAuthUseCase_GetUser_NotFound(t *testing.T) {
	uc, _, _, _ := newTestAuthUseCase()

	q := queries.GetUserQuery{ID: 999}
	_, err := uc.GetUser(context.Background(), q)
	if err == nil {
		t.Fatalf("expected error for not found user")
	}
}

func TestAuthUseCase_ListUsers_Success(t *testing.T) {
	uc, userRepo, _, _ := newTestAuthUseCase()
	seedTestUser(t, userRepo, "user1@example.com", "password123", 1)
	seedTestUser(t, userRepo, "user2@example.com", "password123", 1)

	q := queries.ListUsersQuery{TenantID: 1}
	users, count, err := uc.ListUsers(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestAuthUseCase_CreateTenant_Success(t *testing.T) {
	uc, _, _, _ := newTestAuthUseCase()

	cmd := commands.CreateTenantCommand{
		Name: "Test Org",
		Slug: "test-org",
		Plan: "pro",
	}
	tenant, err := uc.CreateTenant(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tenant.Name != "Test Org" {
		t.Errorf("expected name Test Org, got %s", tenant.Name)
	}
	if tenant.Plan != "pro" {
		t.Errorf("expected plan pro, got %s", tenant.Plan)
	}
}

func TestAuthUseCase_UpdateTenant_Success(t *testing.T) {
	uc, _, tenantRepo, _ := newTestAuthUseCase()
	tenant := &entities.Tenant{Name: "Org", Slug: "org", Plan: "free", Status: "active"}
	tenantRepo.Create(context.Background(), tenant)

	cmd := commands.UpdateTenantCommand{
		ID:   tenant.ID,
		Name: "Updated Org",
		Plan: "enterprise",
	}
	updated, err := uc.UpdateTenant(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "Updated Org" {
		t.Errorf("expected name Updated Org, got %s", updated.Name)
	}
	if updated.Plan != "enterprise" {
		t.Errorf("expected plan enterprise, got %s", updated.Plan)
	}
}

func TestAuthUseCase_DeleteTenant_Success(t *testing.T) {
	uc, _, tenantRepo, _ := newTestAuthUseCase()
	tenant := &entities.Tenant{Name: "Org", Slug: "org", Plan: "free", Status: "active"}
	tenantRepo.Create(context.Background(), tenant)

	err := uc.DeleteTenant(context.Background(), tenant.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = tenantRepo.FindByID(context.Background(), tenant.ID)
	if err == nil {
		t.Errorf("expected error after deletion")
	}
}

func TestAuthUseCase_CreateRole_Success(t *testing.T) {
	uc, _, _, _ := newTestAuthUseCase()

	cmd := commands.CreateRoleCommand{
		TenantID:    1,
		Name:        "admin",
		Description: "Administrator role",
	}
	role, err := uc.CreateRole(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if role.Name != "admin" {
		t.Errorf("expected name admin, got %s", role.Name)
	}
}

func TestAuthUseCase_DeleteRole_Success(t *testing.T) {
	uc, _, _, roleRepo := newTestAuthUseCase()
	role := &entities.Role{TenantID: 1, Name: "admin"}
	roleRepo.Create(context.Background(), role)

	err := uc.DeleteRole(context.Background(), role.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = roleRepo.FindByID(context.Background(), role.ID)
	if err == nil {
		t.Errorf("expected error after deletion")
	}
}
