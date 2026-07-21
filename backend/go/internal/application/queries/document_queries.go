package queries

type GetDocumentQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListDocumentsQuery struct {
	TenantID int64 `json:"tenant_id"`
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
}

type GetDocumentChunksQuery struct {
	DocumentID int64 `json:"document_id" validate:"required"`
}

type GetDocumentSetQuery struct {
	ID int64 `json:"id" validate:"required"`
}

type ListDocumentSetsQuery struct {
	TenantID int64 `json:"tenant_id"`
}
