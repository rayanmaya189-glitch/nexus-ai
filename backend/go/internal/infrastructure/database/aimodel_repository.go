package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAIModelRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAIModelRepository(pool *pgxpool.Pool) *PostgresAIModelRepository {
	return &PostgresAIModelRepository{pool: pool}
}

func (r *PostgresAIModelRepository) FindByID(ctx context.Context, id int64) (*entities.AIModel, error) {
	model := &entities.AIModel{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, model_id, category, size_bytes, parameters, context_length, capabilities, status, max_concurrency, created_at
		 FROM ai_models WHERE id = $1`, id,
	).Scan(
		&model.ID, &model.Name, &model.ModelID, &model.Category,
		&model.SizeBytes, &model.Parameters, &model.ContextLength,
		&model.Capabilities, &model.Status, &model.MaxConcurrency, &model.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (r *PostgresAIModelRepository) FindByCategory(ctx context.Context, category string) ([]*entities.AIModel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, model_id, category, size_bytes, parameters, context_length, capabilities, status, max_concurrency, created_at
		 FROM ai_models WHERE category = $1 ORDER BY name ASC`, category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*entities.AIModel
	for rows.Next() {
		model := &entities.AIModel{}
		err := rows.Scan(
			&model.ID, &model.Name, &model.ModelID, &model.Category,
			&model.SizeBytes, &model.Parameters, &model.ContextLength,
			&model.Capabilities, &model.Status, &model.MaxConcurrency, &model.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

func (r *PostgresAIModelRepository) List(ctx context.Context) ([]*entities.AIModel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, model_id, category, size_bytes, parameters, context_length, capabilities, status, max_concurrency, created_at
		 FROM ai_models ORDER BY category, name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*entities.AIModel
	for rows.Next() {
		model := &entities.AIModel{}
		err := rows.Scan(
			&model.ID, &model.Name, &model.ModelID, &model.Category,
			&model.SizeBytes, &model.Parameters, &model.ContextLength,
			&model.Capabilities, &model.Status, &model.MaxConcurrency, &model.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

func (r *PostgresAIModelRepository) Create(ctx context.Context, model *entities.AIModel) error {
	model.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO ai_models (name, model_id, category, size_bytes, parameters, context_length, capabilities, status, max_concurrency, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id`,
		model.Name, model.ModelID, model.Category, model.SizeBytes,
		model.Parameters, model.ContextLength, model.Capabilities,
		model.Status, model.MaxConcurrency, model.CreatedAt,
	).Scan(&model.ID)
}

func (r *PostgresAIModelRepository) Update(ctx context.Context, model *entities.AIModel) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE ai_models SET name = $1, model_id = $2, category = $3, size_bytes = $4,
		        parameters = $5, context_length = $6, capabilities = $7, status = $8, max_concurrency = $9
		 WHERE id = $10`,
		model.Name, model.ModelID, model.Category, model.SizeBytes,
		model.Parameters, model.ContextLength, model.Capabilities,
		model.Status, model.MaxConcurrency, model.ID,
	)
	return err
}
