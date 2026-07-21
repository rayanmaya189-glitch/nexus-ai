package repositories

import (
	"context"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
)

type DocumentRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.Document, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Document, error)
	Create(ctx context.Context, doc *entities.Document) error
	Update(ctx context.Context, doc *entities.Document) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context, tenantID int64) (int64, error)
}

type DocumentChunkRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.DocumentChunk, error)
	FindByDocumentID(ctx context.Context, documentID int64) ([]*entities.DocumentChunk, error)
	Create(ctx context.Context, chunk *entities.DocumentChunk) error
	CreateBatch(ctx context.Context, chunks []*entities.DocumentChunk) error
	DeleteByDocumentID(ctx context.Context, documentID int64) error
}

type DocumentSetRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.DocumentSet, error)
	FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.DocumentSet, error)
	Create(ctx context.Context, set *entities.DocumentSet) error
	Update(ctx context.Context, set *entities.DocumentSet) error
	Delete(ctx context.Context, id int64) error
}
