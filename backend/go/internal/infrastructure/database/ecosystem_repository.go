package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresIntegrationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresIntegrationRepository(pool *pgxpool.Pool) *PostgresIntegrationRepository {
	return &PostgresIntegrationRepository{pool: pool}
}

func (r *PostgresIntegrationRepository) FindByID(ctx context.Context, id int64) (*entities.Integration, error) {
	intg := &entities.Integration{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, type, config, status, credentials_ref, last_health_check, created_at, updated_at
		 FROM integrations WHERE id = $1`, id,
	).Scan(
		&intg.ID, &intg.TenantID, &intg.Name, &intg.Type, &intg.Config,
		&intg.Status, &intg.CredentialsRef, &intg.LastHealthCheck,
		&intg.CreatedAt, &intg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return intg, nil
}

func (r *PostgresIntegrationRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Integration, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, type, config, status, credentials_ref, last_health_check, created_at, updated_at
		 FROM integrations WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []*entities.Integration
	for rows.Next() {
		intg := &entities.Integration{}
		err := rows.Scan(
			&intg.ID, &intg.TenantID, &intg.Name, &intg.Type, &intg.Config,
			&intg.Status, &intg.CredentialsRef, &intg.LastHealthCheck,
			&intg.CreatedAt, &intg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, intg)
	}
	return integrations, nil
}

func (r *PostgresIntegrationRepository) FindByType(ctx context.Context, tenantID int64, integrationType string) ([]*entities.Integration, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, type, config, status, credentials_ref, last_health_check, created_at, updated_at
		 FROM integrations WHERE tenant_id = $1 AND type = $2 ORDER BY created_at DESC`, tenantID, integrationType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []*entities.Integration
	for rows.Next() {
		intg := &entities.Integration{}
		err := rows.Scan(
			&intg.ID, &intg.TenantID, &intg.Name, &intg.Type, &intg.Config,
			&intg.Status, &intg.CredentialsRef, &intg.LastHealthCheck,
			&intg.CreatedAt, &intg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, intg)
	}
	return integrations, nil
}

func (r *PostgresIntegrationRepository) Create(ctx context.Context, integration *entities.Integration) error {
	now := time.Now()
	integration.CreatedAt = now
	integration.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO integrations (tenant_id, name, type, config, status, credentials_ref, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		integration.TenantID, integration.Name, integration.Type, integration.Config,
		integration.Status, integration.CredentialsRef,
		integration.CreatedAt, integration.UpdatedAt,
	).Scan(&integration.ID)
}

func (r *PostgresIntegrationRepository) Update(ctx context.Context, integration *entities.Integration) error {
	integration.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE integrations SET name = $1, type = $2, config = $3, status = $4,
		        credentials_ref = $5, last_health_check = $6, updated_at = $7
		 WHERE id = $8`,
		integration.Name, integration.Type, integration.Config, integration.Status,
		integration.CredentialsRef, integration.LastHealthCheck,
		integration.UpdatedAt, integration.ID,
	)
	return err
}

func (r *PostgresIntegrationRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM integrations WHERE id = $1`, id)
	return err
}

type PostgresMCPToolRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMCPToolRepository(pool *pgxpool.Pool) *PostgresMCPToolRepository {
	return &PostgresMCPToolRepository{pool: pool}
}

func (r *PostgresMCPToolRepository) FindByID(ctx context.Context, id int64) (*entities.MCPTool, error) {
	tool := &entities.MCPTool{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, description, input_schema, integration_id, status, created_at
		 FROM mcp_tools WHERE id = $1`, id,
	).Scan(
		&tool.ID, &tool.TenantID, &tool.Name, &tool.Description,
		&tool.InputSchema, &tool.IntegrationID, &tool.Status, &tool.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tool, nil
}

func (r *PostgresMCPToolRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.MCPTool, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, description, input_schema, integration_id, status, created_at
		 FROM mcp_tools WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []*entities.MCPTool
	for rows.Next() {
		tool := &entities.MCPTool{}
		err := rows.Scan(
			&tool.ID, &tool.TenantID, &tool.Name, &tool.Description,
			&tool.InputSchema, &tool.IntegrationID, &tool.Status, &tool.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

func (r *PostgresMCPToolRepository) FindByIntegrationID(ctx context.Context, integrationID int64) ([]*entities.MCPTool, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, description, input_schema, integration_id, status, created_at
		 FROM mcp_tools WHERE integration_id = $1 ORDER BY created_at DESC`, integrationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []*entities.MCPTool
	for rows.Next() {
		tool := &entities.MCPTool{}
		err := rows.Scan(
			&tool.ID, &tool.TenantID, &tool.Name, &tool.Description,
			&tool.InputSchema, &tool.IntegrationID, &tool.Status, &tool.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

func (r *PostgresMCPToolRepository) GetEnabledTools(ctx context.Context, tenantID int64) ([]*entities.MCPTool, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, description, input_schema, integration_id, status, created_at
		 FROM mcp_tools WHERE tenant_id = $1 AND status = 'active' ORDER BY name ASC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []*entities.MCPTool
	for rows.Next() {
		tool := &entities.MCPTool{}
		err := rows.Scan(
			&tool.ID, &tool.TenantID, &tool.Name, &tool.Description,
			&tool.InputSchema, &tool.IntegrationID, &tool.Status, &tool.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

func (r *PostgresMCPToolRepository) Create(ctx context.Context, tool *entities.MCPTool) error {
	tool.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO mcp_tools (tenant_id, name, description, input_schema, integration_id, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		tool.TenantID, tool.Name, tool.Description, tool.InputSchema,
		tool.IntegrationID, tool.Status, tool.CreatedAt,
	).Scan(&tool.ID)
}

func (r *PostgresMCPToolRepository) Update(ctx context.Context, tool *entities.MCPTool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE mcp_tools SET name = $1, description = $2, input_schema = $3,
		        integration_id = $4, status = $5
		 WHERE id = $6`,
		tool.Name, tool.Description, tool.InputSchema,
		tool.IntegrationID, tool.Status, tool.ID,
	)
	return err
}

func (r *PostgresMCPToolRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM mcp_tools WHERE id = $1`, id)
	return err
}

type PostgresMCPToolInvocationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMCPToolInvocationRepository(pool *pgxpool.Pool) *PostgresMCPToolInvocationRepository {
	return &PostgresMCPToolInvocationRepository{pool: pool}
}

func (r *PostgresMCPToolInvocationRepository) FindByID(ctx context.Context, id int64) (*entities.MCPToolInvocation, error) {
	inv := &entities.MCPToolInvocation{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tool_id, tenant_id, user_id, input, output, status, latency_ms, created_at
		 FROM mcp_tool_invocations WHERE id = $1`, id,
	).Scan(
		&inv.ID, &inv.ToolID, &inv.TenantID, &inv.UserID,
		&inv.Input, &inv.Output, &inv.Status, &inv.LatencyMs, &inv.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *PostgresMCPToolInvocationRepository) FindByToolID(ctx context.Context, toolID int64) ([]*entities.MCPToolInvocation, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tool_id, tenant_id, user_id, input, output, status, latency_ms, created_at
		 FROM mcp_tool_invocations WHERE tool_id = $1 ORDER BY created_at DESC`, toolID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invocations []*entities.MCPToolInvocation
	for rows.Next() {
		inv := &entities.MCPToolInvocation{}
		err := rows.Scan(
			&inv.ID, &inv.ToolID, &inv.TenantID, &inv.UserID,
			&inv.Input, &inv.Output, &inv.Status, &inv.LatencyMs, &inv.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		invocations = append(invocations, inv)
	}
	return invocations, nil
}

func (r *PostgresMCPToolInvocationRepository) Create(ctx context.Context, invocation *entities.MCPToolInvocation) error {
	invocation.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO mcp_tool_invocations (tool_id, tenant_id, user_id, input, output, status, latency_ms, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		invocation.ToolID, invocation.TenantID, invocation.UserID,
		invocation.Input, invocation.Output, invocation.Status,
		invocation.LatencyMs, invocation.CreatedAt,
	).Scan(&invocation.ID)
}
