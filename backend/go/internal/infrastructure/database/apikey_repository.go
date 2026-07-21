package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAPIKeyRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAPIKeyRepository(pool *pgxpool.Pool) *PostgresAPIKeyRepository {
	return &PostgresAPIKeyRepository{pool: pool}
}

func (r *PostgresAPIKeyRepository) FindByID(ctx context.Context, id int64) (*entities.APIKey, error) {
	key := &entities.APIKey{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, name, key_hash, permissions, expires_at, last_used_at, status, created_at
		 FROM api_keys WHERE id = $1`, id,
	).Scan(
		&key.ID, &key.TenantID, &key.UserID, &key.Name, &key.KeyHash,
		&key.Permissions, &key.ExpiresAt, &key.LastUsedAt, &key.Status, &key.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (r *PostgresAPIKeyRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, user_id, name, key_hash, permissions, expires_at, last_used_at, status, created_at
		 FROM api_keys WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*entities.APIKey
	for rows.Next() {
		key := &entities.APIKey{}
		err := rows.Scan(
			&key.ID, &key.TenantID, &key.UserID, &key.Name, &key.KeyHash,
			&key.Permissions, &key.ExpiresAt, &key.LastUsedAt, &key.Status, &key.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (r *PostgresAPIKeyRepository) FindByUserID(ctx context.Context, userID int64) ([]*entities.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, user_id, name, key_hash, permissions, expires_at, last_used_at, status, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*entities.APIKey
	for rows.Next() {
		key := &entities.APIKey{}
		err := rows.Scan(
			&key.ID, &key.TenantID, &key.UserID, &key.Name, &key.KeyHash,
			&key.Permissions, &key.ExpiresAt, &key.LastUsedAt, &key.Status, &key.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (r *PostgresAPIKeyRepository) Create(ctx context.Context, apiKey *entities.APIKey) error {
	apiKey.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO api_keys (tenant_id, user_id, name, key_hash, permissions, expires_at, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		apiKey.TenantID, apiKey.UserID, apiKey.Name, apiKey.KeyHash,
		apiKey.Permissions, apiKey.ExpiresAt, apiKey.Status, apiKey.CreatedAt,
	).Scan(&apiKey.ID)
}

func (r *PostgresAPIKeyRepository) Update(ctx context.Context, apiKey *entities.APIKey) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET name = $1, permissions = $2, expires_at = $3, last_used_at = $4, status = $5
		 WHERE id = $6`,
		apiKey.Name, apiKey.Permissions, apiKey.ExpiresAt, apiKey.LastUsedAt, apiKey.Status, apiKey.ID,
	)
	return err
}

func (r *PostgresAPIKeyRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM api_keys WHERE id = $1`, id)
	return err
}
