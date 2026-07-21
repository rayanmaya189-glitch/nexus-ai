package usecases

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/domain/repositories"
	"github.com/aeroxe/nexus-backend/pkg/auth"
	"github.com/aeroxe/nexus-backend/pkg/errors"
)

type AuthUseCase struct {
	userRepo  repositories.UserRepository
	tenantRepo repositories.TenantRepository
	roleRepo  repositories.RoleRepository
	jwtManager *auth.JWTManager
}

func NewAuthUseCase(
	userRepo repositories.UserRepository,
	tenantRepo repositories.TenantRepository,
	roleRepo repositories.RoleRepository,
	jwtManager *auth.JWTManager,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		roleRepo:   roleRepo,
		jwtManager: jwtManager,
	}
}

func (uc *AuthUseCase) Login(ctx context.Context, cmd commands.LoginCommand) (*auth.TokenPair, error) {
	user, err := uc.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, errors.Unauthorized("Invalid credentials")
	}

	if user.PasswordHash == "" || !auth.CheckPassword(cmd.Password, user.PasswordHash) {
		return nil, errors.Unauthorized("Invalid credentials")
	}

	if user.Status != "active" {
		return nil, errors.Forbidden("Account is not active")
	}

	now := time.Now()
	user.LastLoginAt = &now
	uc.userRepo.Update(ctx, user)

	return uc.jwtManager.GenerateTokenPair(
		user.ID, user.TenantID, user.Email,
		[]string{"user"},
		[]string{"read", "write"},
	)
}

func (uc *AuthUseCase) Register(ctx context.Context, cmd commands.RegisterCommand) (*entities.User, error) {
	existing, _ := uc.userRepo.FindByEmail(ctx, cmd.Email)
	if existing != nil {
		return nil, errors.Conflict("Email already registered")
	}

	hashedPassword, err := auth.HashPassword(cmd.Password)
	if err != nil {
		return nil, errors.Internal("Failed to hash password")
	}

	user := &entities.User{
		TenantID:     cmd.TenantID,
		Email:        cmd.Email,
		PasswordHash: hashedPassword,
		Name:         cmd.Name,
		Status:       "active",
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, errors.Internal("Failed to create user")
	}

	return user, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, cmd commands.RefreshTokenCommand) (*auth.TokenPair, error) {
	claims, err := uc.jwtManager.ValidateRefreshToken(cmd.RefreshToken)
	if err != nil {
		return nil, errors.Unauthorized("Invalid refresh token")
	}

	user, err := uc.userRepo.FindByEmail(ctx, claims.Subject)
	if err != nil {
		return nil, errors.Unauthorized("User not found")
	}

	return uc.jwtManager.GenerateTokenPair(
		user.ID, user.TenantID, user.Email,
		[]string{"user"},
		[]string{"read", "write"},
	)
}

func (uc *AuthUseCase) GetUser(ctx context.Context, q queries.GetUserQuery) (*entities.User, error) {
	user, err := uc.userRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.NotFound("User not found")
	}
	return user, nil
}

func (uc *AuthUseCase) ListUsers(ctx context.Context, q queries.ListUsersQuery) ([]*entities.User, int64, error) {
	users, err := uc.userRepo.FindByTenantID(ctx, q.TenantID)
	if err != nil {
		return nil, 0, errors.Internal("Failed to list users")
	}
	count, _ := uc.userRepo.Count(ctx, q.TenantID)
	return users, count, nil
}

func (uc *AuthUseCase) CreateUser(ctx context.Context, cmd commands.CreateUserCommand) (*entities.User, error) {
	hashedPassword, err := auth.HashPassword(cmd.Password)
	if err != nil {
		return nil, errors.Internal("Failed to hash password")
	}

	user := &entities.User{
		TenantID:     cmd.TenantID,
		Email:        cmd.Email,
		PasswordHash: hashedPassword,
		Name:         cmd.Name,
		Status:       "active",
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, errors.Internal("Failed to create user")
	}

	return user, nil
}

func (uc *AuthUseCase) UpdateUser(ctx context.Context, cmd commands.UpdateUserCommand) (*entities.User, error) {
	user, err := uc.userRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, errors.NotFound("User not found")
	}

	if cmd.Name != "" {
		user.Name = cmd.Name
	}
	if cmd.Status != "" {
		user.Status = cmd.Status
	}

	user.UpdatedAt = time.Now()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, errors.Internal("Failed to update user")
	}

	return user, nil
}

func (uc *AuthUseCase) DeleteUser(ctx context.Context, id int64) error {
	_, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return errors.NotFound("User not found")
	}
	return uc.userRepo.Delete(ctx, id)
}

func (uc *AuthUseCase) GetTenant(ctx context.Context, q queries.GetTenantQuery) (*entities.Tenant, error) {
	tenant, err := uc.tenantRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.NotFound("Tenant not found")
	}
	return tenant, nil
}

func (uc *AuthUseCase) ListTenants(ctx context.Context, q queries.ListTenantsQuery) ([]*entities.Tenant, int64, error) {
	tenants, total, err := uc.tenantRepo.List(ctx, q.Page, q.PerPage)
	if err != nil {
		return nil, 0, errors.Internal("Failed to list tenants")
	}
	return tenants, total, nil
}

func (uc *AuthUseCase) CreateTenant(ctx context.Context, cmd commands.CreateTenantCommand) (*entities.Tenant, error) {
	tenant := &entities.Tenant{
		Name:   cmd.Name,
		Slug:   cmd.Slug,
		Plan:   cmd.Plan,
		Status: "active",
	}

	if tenant.Plan == "" {
		tenant.Plan = "free"
	}

	if err := uc.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, errors.Internal("Failed to create tenant")
	}

	return tenant, nil
}

func (uc *AuthUseCase) UpdateTenant(ctx context.Context, cmd commands.UpdateTenantCommand) (*entities.Tenant, error) {
	tenant, err := uc.tenantRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, errors.NotFound("Tenant not found")
	}

	if cmd.Name != "" {
		tenant.Name = cmd.Name
	}
	if cmd.Plan != "" {
		tenant.Plan = cmd.Plan
	}
	if cmd.Status != "" {
		tenant.Status = cmd.Status
	}
	if cmd.Settings != "" {
		tenant.Settings = cmd.Settings
	}

	tenant.UpdatedAt = time.Now()
	if err := uc.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, errors.Internal("Failed to update tenant")
	}

	return tenant, nil
}

func (uc *AuthUseCase) DeleteTenant(ctx context.Context, id int64) error {
	_, err := uc.tenantRepo.FindByID(ctx, id)
	if err != nil {
		return errors.NotFound("Tenant not found")
	}
	return uc.tenantRepo.Delete(ctx, id)
}

func (uc *AuthUseCase) GetRole(ctx context.Context, q queries.GetRoleQuery) (*entities.Role, error) {
	role, err := uc.roleRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.NotFound("Role not found")
	}
	return role, nil
}

func (uc *AuthUseCase) ListRoles(ctx context.Context, q queries.ListRolesQuery) ([]*entities.Role, error) {
	roles, err := uc.roleRepo.FindByTenantID(ctx, q.TenantID)
	if err != nil {
		return nil, errors.Internal("Failed to list roles")
	}
	return roles, nil
}

func (uc *AuthUseCase) CreateRole(ctx context.Context, cmd commands.CreateRoleCommand) (*entities.Role, error) {
	role := &entities.Role{
		TenantID:    cmd.TenantID,
		Name:        cmd.Name,
		Description: cmd.Description,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, errors.Internal("Failed to create role")
	}

	return role, nil
}

func (uc *AuthUseCase) DeleteRole(ctx context.Context, id int64) error {
	_, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		return errors.NotFound("Role not found")
	}
	return uc.roleRepo.Delete(ctx, id)
}

func (uc *AuthUseCase) AssignRole(ctx context.Context, cmd commands.AssignRoleCommand) error {
	if err := uc.roleRepo.AssignRole(ctx, cmd.UserID, cmd.RoleID); err != nil {
		return errors.Internal("Failed to assign role")
	}
	return nil
}

func (uc *AuthUseCase) CheckPermission(ctx context.Context, q queries.CheckPermissionQuery) (bool, error) {
	user, err := uc.userRepo.FindByID(ctx, q.UserID)
	if err != nil {
		return false, errors.NotFound("User not found")
	}

	_ = user
	return true, nil
}
