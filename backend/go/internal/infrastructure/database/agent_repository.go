package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAgentRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAgentRepository(pool *pgxpool.Pool) *PostgresAgentRepository {
	return &PostgresAgentRepository{pool: pool}
}

func (r *PostgresAgentRepository) FindByID(ctx context.Context, id int64) (*entities.Agent, error) {
	agent := &entities.Agent{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, agent_type, model, system_prompt, capabilities, status, created_at, updated_at
		 FROM agents WHERE id = $1`, id,
	).Scan(
		&agent.ID, &agent.TenantID, &agent.Name, &agent.AgentType,
		&agent.Model, &agent.SystemPrompt, &agent.Capabilities,
		&agent.Status, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (r *PostgresAgentRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Agent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, agent_type, model, system_prompt, capabilities, status, created_at, updated_at
		 FROM agents WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*entities.Agent
	for rows.Next() {
		agent := &entities.Agent{}
		err := rows.Scan(
			&agent.ID, &agent.TenantID, &agent.Name, &agent.AgentType,
			&agent.Model, &agent.SystemPrompt, &agent.Capabilities,
			&agent.Status, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

func (r *PostgresAgentRepository) FindByType(ctx context.Context, tenantID int64, agentType string) ([]*entities.Agent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, agent_type, model, system_prompt, capabilities, status, created_at, updated_at
		 FROM agents WHERE tenant_id = $1 AND agent_type = $2 ORDER BY created_at DESC`, tenantID, agentType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*entities.Agent
	for rows.Next() {
		agent := &entities.Agent{}
		err := rows.Scan(
			&agent.ID, &agent.TenantID, &agent.Name, &agent.AgentType,
			&agent.Model, &agent.SystemPrompt, &agent.Capabilities,
			&agent.Status, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

func (r *PostgresAgentRepository) Create(ctx context.Context, agent *entities.Agent) error {
	now := time.Now()
	agent.CreatedAt = now
	agent.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO agents (tenant_id, name, agent_type, model, system_prompt, capabilities, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		agent.TenantID, agent.Name, agent.AgentType, agent.Model,
		agent.SystemPrompt, agent.Capabilities, agent.Status,
		agent.CreatedAt, agent.UpdatedAt,
	).Scan(&agent.ID)
}

func (r *PostgresAgentRepository) Update(ctx context.Context, agent *entities.Agent) error {
	agent.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE agents SET name = $1, agent_type = $2, model = $3, system_prompt = $4,
		        capabilities = $5, status = $6, updated_at = $7
		 WHERE id = $8`,
		agent.Name, agent.AgentType, agent.Model, agent.SystemPrompt,
		agent.Capabilities, agent.Status, agent.UpdatedAt, agent.ID,
	)
	return err
}

func (r *PostgresAgentRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM agents WHERE id = $1`, id)
	return err
}
