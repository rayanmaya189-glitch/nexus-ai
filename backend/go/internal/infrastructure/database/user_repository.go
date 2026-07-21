package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id int64) (*entities.User, error) {
	user := &entities.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, email, password_hash, name, status, avatar,
		        last_login_at, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash,
		&user.Name, &user.Status, &user.Avatar,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	user := &entities.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, email, password_hash, name, status, avatar,
		        last_login_at, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash,
		&user.Name, &user.Status, &user.Avatar,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, email, password_hash, name, status, avatar,
		        last_login_at, created_at, updated_at
		 FROM users WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.TenantID, &user.Email, &user.PasswordHash,
			&user.Name, &user.Status, &user.Avatar,
			&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entities.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO users (tenant_id, email, password_hash, name, status, avatar, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		user.TenantID, user.Email, user.PasswordHash, user.Name,
		user.Status, user.Avatar, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *entities.User) error {
	user.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE users SET name = $1, status = $2, avatar = $3, last_login_at = $4, updated_at = $5
		 WHERE id = $6`,
		user.Name, user.Status, user.Avatar, user.LastLoginAt, user.UpdatedAt, user.ID,
	)
	return err
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *PostgresUserRepository) Count(ctx context.Context, tenantID int64) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE tenant_id = $1`, tenantID,
	).Scan(&count)
	return count, err
}
