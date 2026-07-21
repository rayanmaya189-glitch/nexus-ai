package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAISessionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAISessionRepository(pool *pgxpool.Pool) *PostgresAISessionRepository {
	return &PostgresAISessionRepository{pool: pool}
}

func (r *PostgresAISessionRepository) FindByID(ctx context.Context, id int64) (*entities.AISession, error) {
	session := &entities.AISession{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, agent_id, model, status, created_at, updated_at
		 FROM ai_sessions WHERE id = $1`, id,
	).Scan(
		&session.ID, &session.TenantID, &session.UserID, &session.AgentID,
		&session.Model, &session.Status, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *PostgresAISessionRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.AISession, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, user_id, agent_id, model, status, created_at, updated_at
		 FROM ai_sessions WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*entities.AISession
	for rows.Next() {
		session := &entities.AISession{}
		err := rows.Scan(
			&session.ID, &session.TenantID, &session.UserID, &session.AgentID,
			&session.Model, &session.Status, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (r *PostgresAISessionRepository) Create(ctx context.Context, session *entities.AISession) error {
	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO ai_sessions (tenant_id, user_id, agent_id, model, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		session.TenantID, session.UserID, session.AgentID,
		session.Model, session.Status, session.CreatedAt, session.UpdatedAt,
	).Scan(&session.ID)
}

func (r *PostgresAISessionRepository) Update(ctx context.Context, session *entities.AISession) error {
	session.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE ai_sessions SET model = $1, status = $2, updated_at = $3 WHERE id = $4`,
		session.Model, session.Status, session.UpdatedAt, session.ID,
	)
	return err
}

type PostgresAIRequestRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAIRequestRepository(pool *pgxpool.Pool) *PostgresAIRequestRepository {
	return &PostgresAIRequestRepository{pool: pool}
}

func (r *PostgresAIRequestRepository) FindByID(ctx context.Context, id int64) (*entities.AIRequest, error) {
	req := &entities.AIRequest{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, session_id, tenant_id, user_id, agent_id, prompt, response, model,
		        prompt_tokens, response_tokens, latency_ms, status, created_at
		 FROM ai_requests WHERE id = $1`, id,
	).Scan(
		&req.ID, &req.SessionID, &req.TenantID, &req.UserID, &req.AgentID,
		&req.Prompt, &req.Response, &req.Model,
		&req.PromptTokens, &req.ResponseTokens, &req.LatencyMs,
		&req.Status, &req.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *PostgresAIRequestRepository) FindBySessionID(ctx context.Context, sessionID int64) ([]*entities.AIRequest, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, session_id, tenant_id, user_id, agent_id, prompt, response, model,
		        prompt_tokens, response_tokens, latency_ms, status, created_at
		 FROM ai_requests WHERE session_id = $1 ORDER BY created_at ASC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*entities.AIRequest
	for rows.Next() {
		req := &entities.AIRequest{}
		err := rows.Scan(
			&req.ID, &req.SessionID, &req.TenantID, &req.UserID, &req.AgentID,
			&req.Prompt, &req.Response, &req.Model,
			&req.PromptTokens, &req.ResponseTokens, &req.LatencyMs,
			&req.Status, &req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *PostgresAIRequestRepository) Create(ctx context.Context, request *entities.AIRequest) error {
	request.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO ai_requests (session_id, tenant_id, user_id, agent_id, prompt, response, model,
		        prompt_tokens, response_tokens, latency_ms, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id`,
		request.SessionID, request.TenantID, request.UserID, request.AgentID,
		request.Prompt, request.Response, request.Model,
		request.PromptTokens, request.ResponseTokens, request.LatencyMs,
		request.Status, request.CreatedAt,
	).Scan(&request.ID)
}
