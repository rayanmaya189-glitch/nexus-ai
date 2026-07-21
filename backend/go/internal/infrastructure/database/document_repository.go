package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDocumentRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDocumentRepository(pool *pgxpool.Pool) *PostgresDocumentRepository {
	return &PostgresDocumentRepository{pool: pool}
}

func (r *PostgresDocumentRepository) FindByID(ctx context.Context, id int64) (*entities.Document, error) {
	doc := &entities.Document{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, title, content, doc_type, metadata, status, chunk_count, created_at, updated_at
		 FROM documents WHERE id = $1`, id,
	).Scan(
		&doc.ID, &doc.TenantID, &doc.Title, &doc.Content, &doc.DocType,
		&doc.Metadata, &doc.Status, &doc.ChunkCount, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *PostgresDocumentRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Document, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, title, content, doc_type, metadata, status, chunk_count, created_at, updated_at
		 FROM documents WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*entities.Document
	for rows.Next() {
		doc := &entities.Document{}
		err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.Title, &doc.Content, &doc.DocType,
			&doc.Metadata, &doc.Status, &doc.ChunkCount, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (r *PostgresDocumentRepository) Create(ctx context.Context, doc *entities.Document) error {
	now := time.Now()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO documents (tenant_id, title, content, doc_type, metadata, status, chunk_count, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		doc.TenantID, doc.Title, doc.Content, doc.DocType, doc.Metadata,
		doc.Status, doc.ChunkCount, doc.CreatedAt, doc.UpdatedAt,
	).Scan(&doc.ID)
}

func (r *PostgresDocumentRepository) Update(ctx context.Context, doc *entities.Document) error {
	doc.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE documents SET title = $1, content = $2, doc_type = $3, metadata = $4,
		        status = $5, chunk_count = $6, updated_at = $7
		 WHERE id = $8`,
		doc.Title, doc.Content, doc.DocType, doc.Metadata,
		doc.Status, doc.ChunkCount, doc.UpdatedAt, doc.ID,
	)
	return err
}

func (r *PostgresDocumentRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM documents WHERE id = $1`, id)
	return err
}

func (r *PostgresDocumentRepository) Count(ctx context.Context, tenantID int64) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM documents WHERE tenant_id = $1`, tenantID,
	).Scan(&count)
	return count, err
}
