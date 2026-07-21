package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresWorkflowRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresWorkflowRepository(pool *pgxpool.Pool) *PostgresWorkflowRepository {
	return &PostgresWorkflowRepository{pool: pool}
}

func (r *PostgresWorkflowRepository) FindByID(ctx context.Context, id int64) (*entities.Workflow, error) {
	wf := &entities.Workflow{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, description, trigger_type, config, status, version, created_at, updated_at
		 FROM workflows WHERE id = $1`, id,
	).Scan(
		&wf.ID, &wf.TenantID, &wf.Name, &wf.Description, &wf.TriggerType,
		&wf.Config, &wf.Status, &wf.Version, &wf.CreatedAt, &wf.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return wf, nil
}

func (r *PostgresWorkflowRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Workflow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, description, trigger_type, config, status, version, created_at, updated_at
		 FROM workflows WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wfs []*entities.Workflow
	for rows.Next() {
		wf := &entities.Workflow{}
		err := rows.Scan(
			&wf.ID, &wf.TenantID, &wf.Name, &wf.Description, &wf.TriggerType,
			&wf.Config, &wf.Status, &wf.Version, &wf.CreatedAt, &wf.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		wfs = append(wfs, wf)
	}
	return wfs, nil
}

func (r *PostgresWorkflowRepository) Create(ctx context.Context, wf *entities.Workflow) error {
	now := time.Now()
	wf.CreatedAt = now
	wf.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO workflows (tenant_id, name, description, trigger_type, config, status, version, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		wf.TenantID, wf.Name, wf.Description, wf.TriggerType, wf.Config,
		wf.Status, wf.Version, wf.CreatedAt, wf.UpdatedAt,
	).Scan(&wf.ID)
}

func (r *PostgresWorkflowRepository) Update(ctx context.Context, wf *entities.Workflow) error {
	wf.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE workflows SET name = $1, description = $2, trigger_type = $3, config = $4,
		        status = $5, version = $6, updated_at = $7
		 WHERE id = $8`,
		wf.Name, wf.Description, wf.TriggerType, wf.Config,
		wf.Status, wf.Version, wf.UpdatedAt, wf.ID,
	)
	return err
}

func (r *PostgresWorkflowRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM workflows WHERE id = $1`, id)
	return err
}

type PostgresWorkflowStepRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresWorkflowStepRepository(pool *pgxpool.Pool) *PostgresWorkflowStepRepository {
	return &PostgresWorkflowStepRepository{pool: pool}
}

func (r *PostgresWorkflowStepRepository) FindByWorkflowID(ctx context.Context, workflowID int64) ([]*entities.WorkflowStep, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, workflow_id, step_number, name, agent_type, action, config, timeout, created_at
		 FROM workflow_steps WHERE workflow_id = $1 ORDER BY step_number ASC`, workflowID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []*entities.WorkflowStep
	for rows.Next() {
		step := &entities.WorkflowStep{}
		err := rows.Scan(
			&step.ID, &step.WorkflowID, &step.StepNumber, &step.Name, &step.AgentType,
			&step.Action, &step.Config, &step.Timeout, &step.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func (r *PostgresWorkflowStepRepository) Create(ctx context.Context, step *entities.WorkflowStep) error {
	step.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO workflow_steps (workflow_id, step_number, name, agent_type, action, config, timeout, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		step.WorkflowID, step.StepNumber, step.Name, step.AgentType,
		step.Action, step.Config, step.Timeout, step.CreatedAt,
	).Scan(&step.ID)
}

func (r *PostgresWorkflowStepRepository) Update(ctx context.Context, step *entities.WorkflowStep) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE workflow_steps SET name = $1, agent_type = $2, action = $3, config = $4, timeout = $5
		 WHERE id = $6`,
		step.Name, step.AgentType, step.Action, step.Config, step.Timeout, step.ID,
	)
	return err
}
