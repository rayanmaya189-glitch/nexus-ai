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
	log := &entities.AuditLog{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, action, resource_type, resource_id, details,
		        ip_address, user_agent, status, created_at
		 FROM audit_logs WHERE id = $1`, id,
	).Scan(
		&log.ID, &log.TenantID, &log.UserID, &log.Action, &log.ResourceType,
		&log.ResourceID, &log.Details, &log.IPAddress, &log.UserAgent,
		&log.Status, &log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func (r *PostgresAuditLogRepository) FindByTenantID(ctx context.Context, tenantID int64, limit, offset int) ([]*entities.AuditLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, user_id, action, resource_type, resource_id, details,
		        ip_address, user_agent, status, created_at
		 FROM audit_logs WHERE tenant_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*entities.AuditLog
	for rows.Next() {
		log := &entities.AuditLog{}
		err := rows.Scan(
			&log.ID, &log.TenantID, &log.UserID, &log.Action, &log.ResourceType,
			&log.ResourceID, &log.Details, &log.IPAddress, &log.UserAgent,
			&log.Status, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *PostgresAuditLogRepository) Create(ctx context.Context, log *entities.AuditLog) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO audit_logs (tenant_id, user_id, action, resource_type, resource_id, details,
		        ip_address, user_agent, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at`,
		log.TenantID, log.UserID, log.Action, log.ResourceType, log.ResourceID,
		log.Details, log.IPAddress, log.UserAgent, log.Status,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *PostgresAuditLogRepository) Count(ctx context.Context, tenantID int64) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1`, tenantID,
	).Scan(&count)
	return count, err
}
