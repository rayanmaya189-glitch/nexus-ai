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

type PostgresDocumentChunkRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDocumentChunkRepository(pool *pgxpool.Pool) *PostgresDocumentChunkRepository {
	return &PostgresDocumentChunkRepository{pool: pool}
}

func (r *PostgresDocumentChunkRepository) FindByID(ctx context.Context, id int64) (*entities.DocumentChunk, error) {
	chunk := &entities.DocumentChunk{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, document_id, tenant_id, content, chunk_index, embedding, metadata, created_at
		 FROM document_chunks WHERE id = $1`, id,
	).Scan(
		&chunk.ID, &chunk.DocumentID, &chunk.TenantID, &chunk.Content,
		&chunk.ChunkIndex, &chunk.Embedding, &chunk.Metadata, &chunk.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return chunk, nil
}

func (r *PostgresDocumentChunkRepository) FindByDocumentID(ctx context.Context, documentID int64) ([]*entities.DocumentChunk, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, document_id, tenant_id, content, chunk_index, embedding, metadata, created_at
		 FROM document_chunks WHERE document_id = $1 ORDER BY chunk_index ASC`, documentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []*entities.DocumentChunk
	for rows.Next() {
		chunk := &entities.DocumentChunk{}
		err := rows.Scan(
			&chunk.ID, &chunk.DocumentID, &chunk.TenantID, &chunk.Content,
			&chunk.ChunkIndex, &chunk.Embedding, &chunk.Metadata, &chunk.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

func (r *PostgresDocumentChunkRepository) Create(ctx context.Context, chunk *entities.DocumentChunk) error {
	chunk.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO document_chunks (document_id, tenant_id, content, chunk_index, embedding, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		chunk.DocumentID, chunk.TenantID, chunk.Content, chunk.ChunkIndex,
		chunk.Embedding, chunk.Metadata, chunk.CreatedAt,
	).Scan(&chunk.ID)
}

func (r *PostgresDocumentChunkRepository) CreateBatch(ctx context.Context, chunks []*entities.DocumentChunk) error {
	for _, chunk := range chunks {
		if err := r.Create(ctx, chunk); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresDocumentChunkRepository) DeleteByDocumentID(ctx context.Context, documentID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM document_chunks WHERE document_id = $1`, documentID)
	return err
}

type PostgresDocumentSetRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDocumentSetRepository(pool *pgxpool.Pool) *PostgresDocumentSetRepository {
	return &PostgresDocumentSetRepository{pool: pool}
}

func (r *PostgresDocumentSetRepository) FindByID(ctx context.Context, id int64) (*entities.DocumentSet, error) {
	set := &entities.DocumentSet{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, description, status, created_at, updated_at
		 FROM document_sets WHERE id = $1`, id,
	).Scan(
		&set.ID, &set.TenantID, &set.Name, &set.Description,
		&set.Status, &set.CreatedAt, &set.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return set, nil
}

func (r *PostgresDocumentSetRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.DocumentSet, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, description, status, created_at, updated_at
		 FROM document_sets WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []*entities.DocumentSet
	for rows.Next() {
		set := &entities.DocumentSet{}
		err := rows.Scan(
			&set.ID, &set.TenantID, &set.Name, &set.Description,
			&set.Status, &set.CreatedAt, &set.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sets = append(sets, set)
	}
	return sets, nil
}

func (r *PostgresDocumentSetRepository) Create(ctx context.Context, set *entities.DocumentSet) error {
	now := time.Now()
	set.CreatedAt = now
	set.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO document_sets (tenant_id, name, description, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		set.TenantID, set.Name, set.Description, set.Status, set.CreatedAt, set.UpdatedAt,
	).Scan(&set.ID)
}

func (r *PostgresDocumentSetRepository) Update(ctx context.Context, set *entities.DocumentSet) error {
	set.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE document_sets SET name = $1, description = $2, status = $3, updated_at = $4
		 WHERE id = $5`,
		set.Name, set.Description, set.Status, set.UpdatedAt, set.ID,
	)
	return err
}

func (r *PostgresDocumentSetRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM document_sets WHERE id = $1`, id)
	return err
}
