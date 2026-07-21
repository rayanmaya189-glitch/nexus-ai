package commands

type CreateDocumentCommand struct {
	TenantID int64  `json:"tenant_id" validate:"required"`
	Title    string `json:"title" validate:"required"`
	Content  string `json:"content" validate:"required"`
	DocType  string `json:"doc_type"`
	Metadata string `json:"metadata"`
}

type UpdateDocumentCommand struct {
	ID       int64  `json:"id" validate:"required"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	DocType  string `json:"doc_type"`
	Metadata string `json:"metadata"`
	Status   string `json:"status"`
}

type IngestDocumentCommand struct {
	TenantID     int64  `json:"tenant_id" validate:"required"`
	Title        string `json:"title" validate:"required"`
	Content      string `json:"content" validate:"required"`
	DocType      string `json:"doc_type"`
	Metadata     string `json:"metadata"`
	ChunkSize    int    `json:"chunk_size"`
	ChunkOverlap int    `json:"chunk_overlap"`
}

type VectorSearchCommand struct {
	TenantID  int64             `json:"tenant_id" validate:"required"`
	Query     string            `json:"query" validate:"required"`
	TopK      int               `json:"top_k"`
	DocType   string            `json:"doc_type"`
	Filters   map[string]string `json:"filters"`
}

type RAGQueryCommand struct {
	TenantID     int64  `json:"tenant_id" validate:"required"`
	UserID       int64  `json:"user_id" validate:"required"`
	Query        string `json:"query" validate:"required"`
	Model        string `json:"model"`
	TopK         int    `json:"top_k"`
	SystemPrompt string `json:"system_prompt"`
}

type CreateDocumentSetCommand struct {
	TenantID    int64  `json:"tenant_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}
