package usecases

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/domain/repositories"
	nexuserrors "github.com/aeroxe/nexus-backend/pkg/errors"
)

type DocumentUseCase struct {
	docRepo      repositories.DocumentRepository
	chunkRepo    repositories.DocumentChunkRepository
	docSetRepo   repositories.DocumentSetRepository
}

func NewDocumentUseCase(
	docRepo repositories.DocumentRepository,
	chunkRepo repositories.DocumentChunkRepository,
	docSetRepo repositories.DocumentSetRepository,
) *DocumentUseCase {
	return &DocumentUseCase{
		docRepo:    docRepo,
		chunkRepo:  chunkRepo,
		docSetRepo: docSetRepo,
	}
}

func (uc *DocumentUseCase) GetDocument(ctx context.Context, q queries.GetDocumentQuery) (*entities.Document, error) {
	doc, err := uc.docRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, nexuserrors.NotFound("Document not found")
	}
	return doc, nil
}

func (uc *DocumentUseCase) ListDocuments(ctx context.Context, q queries.ListDocumentsQuery) ([]*entities.Document, error) {
	return uc.docRepo.FindByTenantID(ctx, q.TenantID)
}

func (uc *DocumentUseCase) CreateDocument(ctx context.Context, cmd commands.CreateDocumentCommand) (*entities.Document, error) {
	doc := &entities.Document{
		TenantID: cmd.TenantID,
		Title:    cmd.Title,
		Content:  cmd.Content,
		DocType:  cmd.DocType,
		Metadata: cmd.Metadata,
		Status:   "active",
	}

	if doc.DocType == "" {
		doc.DocType = "text"
	}

	if err := uc.docRepo.Create(ctx, doc); err != nil {
		return nil, nexuserrors.Internal("Failed to create document")
	}
	return doc, nil
}

func (uc *DocumentUseCase) UpdateDocument(ctx context.Context, cmd commands.UpdateDocumentCommand) (*entities.Document, error) {
	doc, err := uc.docRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, nexuserrors.NotFound("Document not found")
	}

	if cmd.Title != "" {
		doc.Title = cmd.Title
	}
	if cmd.Content != "" {
		doc.Content = cmd.Content
	}
	if cmd.DocType != "" {
		doc.DocType = cmd.DocType
	}
	if cmd.Metadata != "" {
		doc.Metadata = cmd.Metadata
	}
	if cmd.Status != "" {
		doc.Status = cmd.Status
	}

	doc.UpdatedAt = time.Now()
	if err := uc.docRepo.Update(ctx, doc); err != nil {
		return nil, nexuserrors.Internal("Failed to update document")
	}
	return doc, nil
}

func (uc *DocumentUseCase) DeleteDocument(ctx context.Context, id int64) error {
	_, err := uc.docRepo.FindByID(ctx, id)
	if err != nil {
		return nexuserrors.NotFound("Document not found")
	}
	return uc.docRepo.Delete(ctx, id)
}

func (uc *DocumentUseCase) GetDocumentChunks(ctx context.Context, q queries.GetDocumentChunksQuery) ([]*entities.DocumentChunk, error) {
	return uc.chunkRepo.FindByDocumentID(ctx, q.DocumentID)
}

func (uc *DocumentUseCase) CreateDocumentSet(ctx context.Context, cmd commands.CreateDocumentSetCommand) (*entities.DocumentSet, error) {
	set := &entities.DocumentSet{
		TenantID:    cmd.TenantID,
		Name:        cmd.Name,
		Description: cmd.Description,
		Status:      "active",
	}

	if err := uc.docSetRepo.Create(ctx, set); err != nil {
		return nil, nexuserrors.Internal("Failed to create document set")
	}
	return set, nil
}

func (uc *DocumentUseCase) ListDocumentSets(ctx context.Context, q queries.ListDocumentSetsQuery) ([]*entities.DocumentSet, error) {
	return uc.docSetRepo.FindByTenantID(ctx, q.TenantID)
}

func (uc *DocumentUseCase) DeleteDocumentSet(ctx context.Context, id int64) error {
	_, err := uc.docSetRepo.FindByID(ctx, id)
	if err != nil {
		return nexuserrors.NotFound("Document set not found")
	}
	return uc.docSetRepo.Delete(ctx, id)
}
