package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPlatformConfigurationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPlatformConfigurationRepository(pool *pgxpool.Pool) *PostgresPlatformConfigurationRepository {
	return &PostgresPlatformConfigurationRepository{pool: pool}
}

func (r *PostgresPlatformConfigurationRepository) FindByID(ctx context.Context, id int64) (*entities.PlatformConfiguration, error) {
	cfg := &entities.PlatformConfiguration{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, config_key, config_value, category, description, data_type, is_sensitive, default_value, created_at, updated_at
		 FROM platform_configurations WHERE id = $1`, id,
	).Scan(
		&cfg.ID, &cfg.ConfigKey, &cfg.ConfigValue, &cfg.Category,
		&cfg.Description, &cfg.DataType, &cfg.IsSensitive, &cfg.DefaultValue,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (r *PostgresPlatformConfigurationRepository) FindByKey(ctx context.Context, key string) (*entities.PlatformConfiguration, error) {
	cfg := &entities.PlatformConfiguration{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, config_key, config_value, category, description, data_type, is_sensitive, default_value, created_at, updated_at
		 FROM platform_configurations WHERE config_key = $1`, key,
	).Scan(
		&cfg.ID, &cfg.ConfigKey, &cfg.ConfigValue, &cfg.Category,
		&cfg.Description, &cfg.DataType, &cfg.IsSensitive, &cfg.DefaultValue,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (r *PostgresPlatformConfigurationRepository) FindByCategory(ctx context.Context, category string) ([]*entities.PlatformConfiguration, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, config_key, config_value, category, description, data_type, is_sensitive, default_value, created_at, updated_at
		 FROM platform_configurations WHERE category = $1 ORDER BY config_key ASC`, category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*entities.PlatformConfiguration
	for rows.Next() {
		cfg := &entities.PlatformConfiguration{}
		err := rows.Scan(
			&cfg.ID, &cfg.ConfigKey, &cfg.ConfigValue, &cfg.Category,
			&cfg.Description, &cfg.DataType, &cfg.IsSensitive, &cfg.DefaultValue,
			&cfg.CreatedAt, &cfg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func (r *PostgresPlatformConfigurationRepository) List(ctx context.Context) ([]*entities.PlatformConfiguration, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, config_key, config_value, category, description, data_type, is_sensitive, default_value, created_at, updated_at
		 FROM platform_configurations ORDER BY category, config_key ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*entities.PlatformConfiguration
	for rows.Next() {
		cfg := &entities.PlatformConfiguration{}
		err := rows.Scan(
			&cfg.ID, &cfg.ConfigKey, &cfg.ConfigValue, &cfg.Category,
			&cfg.Description, &cfg.DataType, &cfg.IsSensitive, &cfg.DefaultValue,
			&cfg.CreatedAt, &cfg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func (r *PostgresPlatformConfigurationRepository) Create(ctx context.Context, config *entities.PlatformConfiguration) error {
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO platform_configurations (config_key, config_value, category, description, data_type, is_sensitive, default_value, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		config.ConfigKey, config.ConfigValue, config.Category, config.Description,
		config.DataType, config.IsSensitive, config.DefaultValue,
		config.CreatedAt, config.UpdatedAt,
	).Scan(&config.ID)
}

func (r *PostgresPlatformConfigurationRepository) Update(ctx context.Context, config *entities.PlatformConfiguration) error {
	config.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE platform_configurations SET config_value = $1, category = $2, description = $3,
		        data_type = $4, is_sensitive = $5, default_value = $6, updated_at = $7
		 WHERE id = $8`,
		config.ConfigValue, config.Category, config.Description,
		config.DataType, config.IsSensitive, config.DefaultValue,
		config.UpdatedAt, config.ID,
	)
	return err
}

func (r *PostgresPlatformConfigurationRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM platform_configurations WHERE id = $1`, id)
	return err
}

type PostgresTenantConfigurationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTenantConfigurationRepository(pool *pgxpool.Pool) *PostgresTenantConfigurationRepository {
	return &PostgresTenantConfigurationRepository{pool: pool}
}

func (r *PostgresTenantConfigurationRepository) FindByID(ctx context.Context, id int64) (*entities.TenantConfiguration, error) {
	cfg := &entities.TenantConfiguration{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, config_key, config_value, created_at, updated_at
		 FROM tenant_configurations WHERE id = $1`, id,
	).Scan(
		&cfg.ID, &cfg.TenantID, &cfg.ConfigKey, &cfg.ConfigValue,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (r *PostgresTenantConfigurationRepository) FindByTenantAndKey(ctx context.Context, tenantID int64, key string) (*entities.TenantConfiguration, error) {
	cfg := &entities.TenantConfiguration{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, config_key, config_value, created_at, updated_at
		 FROM tenant_configurations WHERE tenant_id = $1 AND config_key = $2`, tenantID, key,
	).Scan(
		&cfg.ID, &cfg.TenantID, &cfg.ConfigKey, &cfg.ConfigValue,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (r *PostgresTenantConfigurationRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.TenantConfiguration, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, config_key, config_value, created_at, updated_at
		 FROM tenant_configurations WHERE tenant_id = $1 ORDER BY config_key ASC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*entities.TenantConfiguration
	for rows.Next() {
		cfg := &entities.TenantConfiguration{}
		err := rows.Scan(
			&cfg.ID, &cfg.TenantID, &cfg.ConfigKey, &cfg.ConfigValue,
			&cfg.CreatedAt, &cfg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func (r *PostgresTenantConfigurationRepository) Create(ctx context.Context, config *entities.TenantConfiguration) error {
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO tenant_configurations (tenant_id, config_key, config_value, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		config.TenantID, config.ConfigKey, config.ConfigValue,
		config.CreatedAt, config.UpdatedAt,
	).Scan(&config.ID)
}

func (r *PostgresTenantConfigurationRepository) Update(ctx context.Context, config *entities.TenantConfiguration) error {
	config.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE tenant_configurations SET config_value = $1, updated_at = $2
		 WHERE id = $3`,
		config.ConfigValue, config.UpdatedAt, config.ID,
	)
	return err
}

func (r *PostgresTenantConfigurationRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tenant_configurations WHERE id = $1`, id)
	return err
}
