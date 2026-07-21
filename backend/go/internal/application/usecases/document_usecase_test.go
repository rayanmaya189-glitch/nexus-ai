package usecases_test

import (
	"context"
	"testing"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/testutil"
)

func newTestDocumentUseCase() (*usecases.DocumentUseCase, *testutil.MockDocumentRepository, *testutil.MockDocumentChunkRepository, *testutil.MockDocumentSetRepository) {
	docRepo := testutil.NewMockDocumentRepository()
	chunkRepo := testutil.NewMockDocumentChunkRepository()
	docSetRepo := testutil.NewMockDocumentSetRepository()
	uc := usecases.NewDocumentUseCase(docRepo, chunkRepo, docSetRepo)
	return uc, docRepo, chunkRepo, docSetRepo
}

func TestDocumentUseCase_CreateDocument_Success(t *testing.T) {
	uc, _, _, _ := newTestDocumentUseCase()

	cmd := commands.CreateDocumentCommand{
		TenantID: 1,
		Title:    "Test Document",
		Content:  "Some content here",
		DocType:  "text",
	}
	doc, err := uc.CreateDocument(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if doc.Title != "Test Document" {
		t.Errorf("expected title Test Document, got %s", doc.Title)
	}
	if doc.Status != "active" {
		t.Errorf("expected status active, got %s", doc.Status)
	}
}

func TestDocumentUseCase_GetDocument_Success(t *testing.T) {
	uc, _, _, _ := newTestDocumentUseCase()
	doc, _ := uc.CreateDocument(context.Background(), commands.CreateDocumentCommand{
		TenantID: 1, Title: "Doc", Content: "Content",
	})

	q := queries.GetDocumentQuery{ID: doc.ID}
	found, err := uc.GetDocument(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Title != "Doc" {
		t.Errorf("expected title Doc, got %s", found.Title)
	}
}

func TestDocumentUseCase_GetDocument_NotFound(t *testing.T) {
	uc, _, _, _ := newTestDocumentUseCase()

	q := queries.GetDocumentQuery{ID: 999}
	_, err := uc.GetDocument(context.Background(), q)
	if err == nil {
		t.Fatalf("expected error for not found document")
	}
}

func TestDocumentUseCase_ListDocuments_Success(t *testing.T) {
	uc, _, _, _ := newTestDocumentUseCase()
	uc.CreateDocument(context.Background(), commands.CreateDocumentCommand{
		TenantID: 1, Title: "Doc 1", Content: "Content 1",
	})
	uc.CreateDocument(context.Background(), commands.CreateDocumentCommand{
		TenantID: 1, Title: "Doc 2", Content: "Content 2",
	})

	q := queries.ListDocumentsQuery{TenantID: 1}
	docs, err := uc.ListDocuments(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs))
	}
}

func TestDocumentUseCase_UpdateDocument_Success(t *testing.T) {
	uc, _, _, _ := newTestDocumentUseCase()
	doc, _ := uc.CreateDocument(context.Background(), commands.CreateDocumentCommand{
		TenantID: 1, Title: "Old Title", Content: "Content",
	})

	cmd := commands.UpdateDocumentCommand{
		ID:    doc.ID,
		Title: "New Title",
	}
	updated, err := uc.UpdateDocument(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected title New Title, got %s", updated.Title)
	}
}

func TestDocumentUseCase_DeleteDocument_Success(t *testing.T) {
	uc, docRepo, _, _ := newTestDocumentUseCase()
	doc, _ := uc.CreateDocument(context.Background(), commands.CreateDocumentCommand{
		TenantID: 1, Title: "Doc", Content: "Content",
	})

	err := uc.DeleteDocument(context.Background(), doc.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = docRepo.FindByID(context.Background(), doc.ID)
	if err == nil {
		t.Errorf("expected error after deletion")
	}
}

func TestDocumentUseCase_DeleteDocument_NotFound(t *testing.T) {
	uc, _, _, _ := newTestDocumentUseCase()

	err := uc.DeleteDocument(context.Background(), 999)
	if err == nil {
		t.Fatalf("expected error for not found document")
	}
}

func TestDocumentUseCase_CreateDocumentSet_Success(t *testing.T) {
	uc, _, _, _ := newTestDocumentUseCase()

	cmd := commands.CreateDocumentSetCommand{
		TenantID:    1,
		Name:        "Knowledge Base",
		Description: "Internal knowledge base",
	}
	set, err := uc.CreateDocumentSet(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if set.Name != "Knowledge Base" {
		t.Errorf("expected name Knowledge Base, got %s", set.Name)
	}
}

func TestDocumentUseCase_ListDocumentSets_Success(t *testing.T) {
	uc, _, _, _ := newTestDocumentUseCase()
	uc.CreateDocumentSet(context.Background(), commands.CreateDocumentSetCommand{
		TenantID: 1, Name: "Set 1",
	})
	uc.CreateDocumentSet(context.Background(), commands.CreateDocumentSetCommand{
		TenantID: 1, Name: "Set 2",
	})

	q := queries.ListDocumentSetsQuery{TenantID: 1}
	sets, err := uc.ListDocumentSets(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(sets) != 2 {
		t.Errorf("expected 2 document sets, got %d", len(sets))
	}
}
