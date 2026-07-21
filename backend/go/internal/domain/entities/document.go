package entities

import "time"

type Document struct {
	ID          int64     `json:"id" db:"id"`
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	DocType     string    `json:"doc_type" db:"doc_type"`
	Metadata    string    `json:"metadata" db:"metadata"`
	Status      string    `json:"status" db:"status"`
	ChunkCount  int       `json:"chunk_count" db:"chunk_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type DocumentChunk struct {
	ID          int64     `json:"id" db:"id"`
	DocumentID  int64     `json:"document_id" db:"document_id"`
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	Content     string    `json:"content" db:"content"`
	ChunkIndex  int       `json:"chunk_index" db:"chunk_index"`
	Embedding   []float32 `json:"-" db:"embedding"`
	Metadata    string    `json:"metadata" db:"metadata"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type DocumentSet struct {
	ID          int64     `json:"id" db:"id"`
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
