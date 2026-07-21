package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRoleRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRoleRepository(pool *pgxpool.Pool) *PostgresRoleRepository {
	return &PostgresRoleRepository{pool: pool}
}

func (r *PostgresRoleRepository) FindByID(ctx context.Context, id int64) (*entities.Role, error) {
	role := &entities.Role{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, description, is_default, created_at
		 FROM roles WHERE id = $1`, id,
	).Scan(
		&role.ID, &role.TenantID, &role.Name, &role.Description,
		&role.IsDefault, &role.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *PostgresRoleRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Role, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, description, is_default, created_at
		 FROM roles WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*entities.Role
	for rows.Next() {
		role := &entities.Role{}
		err := rows.Scan(
			&role.ID, &role.TenantID, &role.Name, &role.Description,
			&role.IsDefault, &role.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *PostgresRoleRepository) Create(ctx context.Context, role *entities.Role) error {
	role.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO roles (tenant_id, name, description, is_default, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		role.TenantID, role.Name, role.Description, role.IsDefault, role.CreatedAt,
	).Scan(&role.ID)
}

func (r *PostgresRoleRepository) Update(ctx context.Context, role *entities.Role) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE roles SET name = $1, description = $2, is_default = $3
		 WHERE id = $4`,
		role.Name, role.Description, role.IsDefault, role.ID,
	)
	return err
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM roles WHERE id = $1`, id)
	return err
}

func (r *PostgresRoleRepository) AssignRole(ctx context.Context, userID, roleID int64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, roleID,
	)
	return err
}

func (r *PostgresRoleRepository) RemoveRole(ctx context.Context, userID, roleID int64) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`,
		userID, roleID,
	)
	return err
}
