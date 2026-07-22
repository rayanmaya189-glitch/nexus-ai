# AeroXe Nexus AI — RAG Module

## Enterprise Knowledge Intelligence: Ingestion, Embeddings, Hybrid Search & Knowledge Graph

> **Modular Monolith Module:** This document describes the `nexus-rag` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces (see [Communication Architecture](12-communication-architecture.md)).

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-rag` |
| Crate | `nexus-rag` (workspace member) |
| Bounded Context | RAG Knowledge Intelligence |
| Domain Type | Core Domain |
| Language | Rust |
| Schema | `rag_` (in shared PostgreSQL + pgvector) |
| Dependencies | Ollama (embeddings), MinIO, Elasticsearch |

---

## 2. Purpose

The RAG module provides enterprise knowledge intelligence within the `aeroxe-nexus` monolith by:

- Ingesting documents from multiple sources (PDF, DOCX, HTML, Markdown, code)
- Parsing and chunking content using semantic chunking
- Generating vector embeddings for semantic search
- Performing hybrid search (vector + keyword + graph)
- Re-ranking results for relevance
- Feeding context to LLM for answer generation
- Enforcing document-level access control

---

## 3. Aggregate Design

### KnowledgeDocument Aggregate

```
KnowledgeDocument (Aggregate Root)
├── Metadata
│   ├── DocumentId
│   ├── TenantId
│   ├── FileName
│   ├── FileType
│   ├── Status
│   ├── CreatedAt
│   └── Tags[]
├── Chunks[]
│   ├── ChunkId
│   ├── Content
│   ├── Position
│   ├── Embedding (vector(768))
│   └── Metadata
└── AccessControl
    ├── Classification
    └── AllowedRoles[]
```

### Entities

| Entity | Attributes |
|---|---|
| Document | DocumentId, TenantId, FileName, Type, Status, Size |
| Chunk | ChunkId, DocumentId, Content, ChunkIndex, Embedding |
| DocumentMetadata | DocumentId, Metadata (JSONB) |

### Value Objects

| Value Object | Type |
|---|---|
| `DocumentId` | i64 |
| `EmbeddingVector` | 768-dimensional float array |
| `DocumentType` | Enum (pdf, docx, html, markdown, code, database) |
| `ChunkStrategy` | Enum (semantic, fixed, recursive) |

---

## 4. Public API Trait

```rust
// nexus-rag/src/interfaces/api.rs
#[async_trait]
pub trait RagService: Send + Sync {
    async fn search(&self, query: SearchQuery) -> Result<SearchResults, RagError>;
    async fn upload_document(&self, req: UploadRequest) -> Result<DocumentStatus, RagError>;
    async fn get_document_status(&self, id: DocumentId) -> Result<Option<DocumentStatus>, RagError>;
    async fn delete_document(&self, id: DocumentId) -> Result<(), RagError>;
}

pub struct SearchQuery {
    pub query: String,
    pub limit: u32,
    pub tenant_id: TenantId,
    pub filters: Vec<String>,
}

pub struct SearchResults {
    pub results: Vec<DocumentResult>,
    pub total_latency_ms: f64,
}
```

> **Note:** Other modules (like `nexus-agent`) call `RagService` methods synchronously via trait dispatch. Document upload processing is async via NATS.

---

## 5. Document Ingestion Pipeline

### Step 1: Document Upload

```
Client -> POST /api/v1/rag/documents (multipart)
    |
    v
Store raw file in MinIO (aeroxe-documents bucket)
    |
    v
Create document record in PostgreSQL
    |
    v
Publish NATS event: aeroxe.v1.rag.document.uploaded
```

### Step 2: Document Processing

```
NATS Consumer receives document.uploaded
    |
    v
Parser (format-specific: PDF, DOCX, HTML, Markdown)
    |
    v
Text Extraction
    |
    v
Cleaning (remove headers, footers, boilerplate)
    |
    v
Chunking (Semantic Chunking Strategy)
    |
    v
Metadata Extraction (department, category, security level)
    |
    v
Embedding Generation (via Ollama embedding model)
    |
    v
Store chunks + embeddings in pgvector
    |
    v
Publish NATS event: aeroxe.v1.rag.document.processed
```

### Step 3: Knowledge Graph Update

```
NATS Consumer receives document.processed
    |
    v
Extract entities and relationships
    |
    v
Update Apache AGE knowledge graph
    |
    v
Publish NATS event: aeroxe.v1.rag.embedding.created
```

---

## 6. Semantic Chunking

Instead of fixed-size chunking, AeroXe uses semantic chunking that preserves meaning:

### Traditional Approach
```
Fixed 500 tokens per chunk
Problem: Breaks in the middle of paragraphs/sections
```

### AeroXe Semantic Approach
```
Respects document structure:
- Sections and subsections
- Paragraph boundaries
- Code blocks
- Tables
- Lists

Output: Chunks that are semantically complete
```

### Chunk Configuration

| Parameter | Value |
|---|---|
| Strategy | Semantic |
| Max Chunk Size | 1024 tokens |
| Min Chunk Size | 100 tokens |
| Overlap | 200 tokens |
| Separator Priority | `\n\n`, `\n`, `. `, ` ` |

---

## 7. Hybrid Search Architecture

The RAG service performs multi-modal search combining four search strategies:

### Search Pipeline

```
User Query
    |
    v
[1] Query Understanding
    |  - Intent classification
    |  - Query expansion
    |  - Synonym resolution
    |
    v
[2] Parallel Search (Fan-out)
    |
    ├── Vector Search (pgvector)
    |   Semantic similarity search using cosine distance
    |   Finds: Similar meaning, context
    |
    ├── Keyword Search (Elasticsearch)
    |   Full-text search with BM25 scoring
    |   Finds: Exact terms, technical references
    |
    ├── Knowledge Graph Search (Apache AGE)
    |   Relationship traversal
    |   Finds: Connected entities, paths
    |
    └── Database Query (PostgreSQL)
        Structured data lookup
        Finds: Facts, statistics, records
    |
    v
[3] Result Fusion
    |  - Merge results from all sources
    |  - Deduplicate
    |  - Score normalization
    |
    v
[4] Re-ranking
    |  - Cross-encoder re-ranking model
    |  - Re-score top 100 results
    |  - Return top 5-10
    |
    v
[5] Security Filtering
    |  - Tenant isolation
    |  - Document-level access control
    |  - Data classification check
    |
    v
[6] Context Assembly
    |  - Format context for LLM
    |  - Include source citations
    |  - Token budget management
    |
    v
[7] LLM Reasoning (Command-R 7B)
    |  - Generate answer from context
    |  - Cite sources
    |  - Confidence scoring
    |
    v
Final Answer
```

---

## 8. Vector Search (pgvector)

### Embedding Configuration

| Parameter | Value |
|---|---|
| Model | nomic-embed-text |
| Dimension | 768 |
| Index Type | IVFFlat |
| Distance Metric | Cosine Similarity |
| Lists | 100 (tunable) |
| Probes | 10 (tunable) |

### Vector Index

```sql
CREATE INDEX embedding_index
ON document_chunks
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);
```

### Search Query

```sql
SELECT id, content, chunk_index,
       1 - (embedding <=> $1) AS similarity
FROM document_chunks dc
JOIN documents d ON dc.document_id = d.id
WHERE d.tenant_id = $2
  AND 1 - (embedding <=> $1) > 0.7
ORDER BY embedding <=> $1
LIMIT 10;
```

---

## 9. Elasticsearch Integration

### Indices

| Index | Purpose |
|---|---|
| `documents` | Full-text document search |
| `audit_events` | Audit log search |
| `ai_logs` | AI conversation logs |
| `security_events` | Security event search |

### Document Index Mapping

```json
{
  "mappings": {
    "properties": {
      "document_id": { "type": "keyword" },
      "tenant_id": { "type": "keyword" },
      "content": { "type": "text", "analyzer": "standard" },
      "title": { "type": "text", "analyzer": "standard" },
      "tags": { "type": "keyword" },
      "created_at": { "type": "date" }
    }
  }
}
```

---

## 10. Knowledge Graph (Apache AGE)

### Graph Entities (Nodes)

| Node Type | Properties |
|---|---|
| Customer | id, name, email, plan |
| Company | id, name, industry |
| Device | id, type, model, location |
| Network | id, name, type |
| Document | id, title, type |
| Agent | id, name, capability |

### Relationships

| Relationship | From | To |
|---|---|---|
| `OWNS` | Customer | Device |
| `CONNECTED_TO` | Device | Network |
| `DEPENDS_ON` | Device | Device |
| `RELATED_TO` | Document | Document |
| `BELONGS_TO` | Device | Company |

### Graph Query Example

```sql
-- Find all customers affected by an OLT issue
SELECT * FROM cypher('knowledge_graph', $$
  MATCH (c:Customer)-[:OWNS]->(d:Device)-[:CONNECTED_TO]->(n:Network {name: 'OLT-Jalgaon-1'})
  RETURN c.name, d.type, d.model
$$) AS (name VARCHAR, device_type VARCHAR, model VARCHAR);
```

---

## 11. Re-ranking System

After initial search, results are re-ranked for precision:

```
100 candidate results (from all search sources)
    |
    v
Cross-Encoder Re-ranking Model
    - Compare query against each result
    - Score relevance (0.0 - 1.0)
    |
    v
Sort by relevance score
    |
    v
Top 5 results selected
    |
    v
Passed to LLM for answer generation
```

---

## 12. RAG Security Layer

Before retrieval, security checks are applied:

| Check | Description |
|---|---|
| Tenant Isolation | Only retrieve documents belonging to user's tenant |
| Agent Scope Isolation | Agent queries filtered to bound document sets only |
| Document Access | User must have read permission on document |
| Data Classification | Support users cannot access financial reports |
| Content Filtering | Filter results based on ABAC policies |

### Agent Scope Isolation

When an agent performs RAG search, results are filtered to its bound document sets:

```sql
-- Get document IDs from agent's bound document sets
WITH agent_scope AS (
    SELECT dsd.document_id
    FROM agent_document_sets ads
    JOIN document_set_documents dsd ON dsd.document_set_id = ads.document_set_id
    WHERE ads.agent_id = $1
      AND ads.tenant_id = $2
)
-- Search only within scoped documents
SELECT dc.id, dc.content, dc.metadata,
       dc.embedding <=> $3 AS distance
FROM document_chunks dc
WHERE dc.document_id IN (SELECT document_id FROM agent_scope)
ORDER BY distance
LIMIT $4;
```

### Isolation Enforcement

| Layer | Enforcement |
|---|---|
| Agent Orchestrator | Scope injected into execution context |
| RAG Service | Query filtered by document set scope |
| Vector Search | pgvector query scoped to bound documents |
| Audit | All access logged with scope info |

---

## 13. Database Schema (rag_db)

### documents

```sql
CREATE TABLE rag.documents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    filename TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'uploaded',
    size_bytes BIGINT,
    storage_path TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP
);
```

### document_chunks

```sql
CREATE TABLE rag.chunks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES rag.documents(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    chunk_index INT NOT NULL,
    token_count INT,
    embedding vector(768),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### document_metadata

```sql
CREATE TABLE rag.document_metadata (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES rag.documents(id) ON DELETE CASCADE,
    metadata JSONB NOT NULL
);
```

---

## 14. REST API Endpoints

### Upload Document

```
POST /api/v1/rag/documents
Content-Type: multipart/form-data
```

**Response:**
```json
{
  "document_id": "uuid",
  "status": "processing"
}
```

### Search Knowledge

```
POST /api/v1/rag/search
```

**Request:**
```json
{
  "query": "How to configure ONU?",
  "limit": 5
}
```

**Response:**
```json
{
  "results": [
    {
      "title": "ONU Configuration Guide",
      "score": 0.91,
      "content": "Step 1: Connect to ONU via...",
      "source": "network-guide.pdf",
      "metadata": {
        "department": "network",
        "category": "troubleshooting"
      }
    }
  ]
}
```

### Get Document Status

```
GET /api/v1/rag/documents/{id}/status
```

---

## 15. NATS Events

### Published

| Subject | Event |
|---|---|
| `aeroxe.v1.rag.document.uploaded` | Document received |
| `aeroxe.v1.rag.document.processed` | Processing complete |
| `aeroxe.v1.rag.embedding.created` | Embeddings stored |
| `aeroxe.v1.rag.knowledge.updated` | Knowledge base modified |

### Subscribed

| Subject | Handler |
|---|---|
| `aeroxe.v1.rag.document.uploaded` | Trigger document processing |

---

## 16. Performance Targets

| Operation | Target |
|---|---|
| Document Upload | < 2s |
| Document Processing | < 30s (per document) |
| Vector Search | < 200ms |
| Hybrid Search | < 500ms |
| Re-ranking | < 100ms |
| Full RAG Answer | < 2s |

---

## 17. MinIO Storage

### Buckets

| Bucket | Purpose |
|---|---|
| `aeroxe-documents` | Uploaded documents |
| `aeroxe-images` | Vision analysis images |
| `aeroxe-model-files` | Custom model files |
| `aeroxe-backups` | System backups |
