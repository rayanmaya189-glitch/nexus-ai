package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresMemoryRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMemoryRepository(pool *pgxpool.Pool) *PostgresMemoryRepository {
	return &PostgresMemoryRepository{pool: pool}
}

func (r *PostgresMemoryRepository) FindByID(ctx context.Context, id int64) (*entities.Memory, error) {
	mem := &entities.Memory{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, agent_id, session_id, memory_type, content, summary,
		        importance, access_count, expires_at, created_at
		 FROM memory_entries WHERE id = $1`, id,
	).Scan(
		&mem.ID, &mem.TenantID, &mem.UserID, &mem.AgentID, &mem.SessionID,
		&mem.MemoryType, &mem.Content, &mem.Summary, &mem.Importance,
		&mem.AccessCount, &mem.ExpiresAt, &mem.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return mem, nil
}

func (r *PostgresMemoryRepository) FindByAgentID(ctx context.Context, agentID int64) ([]*entities.Memory, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, user_id, agent_id, session_id, memory_type, content, summary,
		        importance, access_count, expires_at, created_at
		 FROM memory_entries WHERE agent_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
		 ORDER BY importance DESC, created_at DESC`, agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*entities.Memory
	for rows.Next() {
		mem := &entities.Memory{}
		err := rows.Scan(
			&mem.ID, &mem.TenantID, &mem.UserID, &mem.AgentID, &mem.SessionID,
			&mem.MemoryType, &mem.Content, &mem.Summary, &mem.Importance,
			&mem.AccessCount, &mem.ExpiresAt, &mem.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		memories = append(memories, mem)
	}
	return memories, nil
}

func (r *PostgresMemoryRepository) FindByTenantAndUser(ctx context.Context, tenantID, userID int64) ([]*entities.Memory, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, user_id, agent_id, session_id, memory_type, content, summary,
		        importance, access_count, expires_at, created_at
		 FROM memory_entries WHERE tenant_id = $1 AND user_id = $2 AND (expires_at IS NULL OR expires_at > NOW())
		 ORDER BY importance DESC, created_at DESC`, tenantID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*entities.Memory
	for rows.Next() {
		mem := &entities.Memory{}
		err := rows.Scan(
			&mem.ID, &mem.TenantID, &mem.UserID, &mem.AgentID, &mem.SessionID,
			&mem.MemoryType, &mem.Content, &mem.Summary, &mem.Importance,
			&mem.AccessCount, &mem.ExpiresAt, &mem.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		memories = append(memories, mem)
	}
	return memories, nil
}

func (r *PostgresMemoryRepository) Create(ctx context.Context, mem *entities.Memory) error {
	mem.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO memory_entries (tenant_id, user_id, agent_id, session_id, memory_type, content, summary,
		        importance, access_count, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id`,
		mem.TenantID, mem.UserID, mem.AgentID, mem.SessionID, mem.MemoryType,
		mem.Content, mem.Summary, mem.Importance, mem.AccessCount, mem.ExpiresAt, mem.CreatedAt,
	).Scan(&mem.ID)
}

func (r *PostgresMemoryRepository) Update(ctx context.Context, mem *entities.Memory) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE memory_entries SET content = $1, summary = $2, importance = $3, access_count = $4, updated_at = NOW()
		 WHERE id = $5`,
		mem.Content, mem.Summary, mem.Importance, mem.AccessCount, mem.ID,
	)
	return err
}

func (r *PostgresMemoryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM memory_entries WHERE id = $1`, id)
	return err
}
