package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTenantRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTenantRepository(pool *pgxpool.Pool) *PostgresTenantRepository {
	return &PostgresTenantRepository{pool: pool}
}

func (r *PostgresTenantRepository) FindByID(ctx context.Context, id int64) (*entities.Tenant, error) {
	tenant := &entities.Tenant{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, plan, status, settings, created_at, updated_at
		 FROM tenants WHERE id = $1`, id,
	).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan,
		&tenant.Status, &tenant.Settings, &tenant.CreatedAt, &tenant.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func (r *PostgresTenantRepository) FindBySlug(ctx context.Context, slug string) (*entities.Tenant, error) {
	tenant := &entities.Tenant{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, plan, status, settings, created_at, updated_at
		 FROM tenants WHERE slug = $1`, slug,
	).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan,
		&tenant.Status, &tenant.Settings, &tenant.CreatedAt, &tenant.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func (r *PostgresTenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	now := time.Now()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO tenants (name, slug, plan, status, settings, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		tenant.Name, tenant.Slug, tenant.Plan, tenant.Status,
		tenant.Settings, tenant.CreatedAt, tenant.UpdatedAt,
	).Scan(&tenant.ID)
}

func (r *PostgresTenantRepository) Update(ctx context.Context, tenant *entities.Tenant) error {
	tenant.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE tenants SET name = $1, plan = $2, status = $3, settings = $4, updated_at = $5
		 WHERE id = $6`,
		tenant.Name, tenant.Plan, tenant.Status, tenant.Settings, tenant.UpdatedAt, tenant.ID,
	)
	return err
}

func (r *PostgresTenantRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, id)
	return err
}

func (r *PostgresTenantRepository) List(ctx context.Context, page, perPage int) ([]*entities.Tenant, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, name, slug, plan, status, settings, created_at, updated_at
		 FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2`, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tenants []*entities.Tenant
	for rows.Next() {
		tenant := &entities.Tenant{}
		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan,
			&tenant.Status, &tenant.Settings, &tenant.CreatedAt, &tenant.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, tenant)
	}
	return tenants, total, nil
}
