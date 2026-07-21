package database

import (
	"context"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAuditLogRepository(pool *pgxpool.Pool) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{pool: pool}
}

func (r *PostgresAuditLogRepository) FindByID(ctx context.Context, id int64) (*entities.AuditLog, error) {
	entry := &entities.AuditLog{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, action, resource_type, resource_id, details,
		        ip_address, user_agent, status, created_at
		 FROM audit_logs WHERE id = $1`, id,
	).Scan(
		&entry.ID, &entry.TenantID, &entry.UserID, &entry.Action, &entry.ResourceType,
		&entry.ResourceID, &entry.Details, &entry.IPAddress, &entry.UserAgent,
		&entry.Status, &entry.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (r *PostgresAuditLogRepository) FindByTenantID(ctx context.Context, tenantID int64, limit, offset int) ([]*entities.AuditLog, int64, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, user_id, action, resource_type, resource_id, details,
		        ip_address, user_agent, status, created_at
		 FROM audit_logs WHERE tenant_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []*entities.AuditLog
	for rows.Next() {
		entry := &entities.AuditLog{}
		err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.UserID, &entry.Action, &entry.ResourceType,
			&entry.ResourceID, &entry.Details, &entry.IPAddress, &entry.UserAgent,
			&entry.Status, &entry.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
	}

	var count int64
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1`, tenantID,
	).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return entries, count, nil
}

func (r *PostgresAuditLogRepository) Create(ctx context.Context, entry *entities.AuditLog) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO audit_logs (tenant_id, user_id, action, resource_type, resource_id, details,
		        ip_address, user_agent, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at`,
		entry.TenantID, entry.UserID, entry.Action, entry.ResourceType, entry.ResourceID,
		entry.Details, entry.IPAddress, entry.UserAgent, entry.Status,
	).Scan(&entry.ID, &entry.CreatedAt)
}

func (r *PostgresAuditLogRepository) Count(ctx context.Context, tenantID int64) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1`, tenantID,
	).Scan(&count)
	return count, err
}
