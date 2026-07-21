# AeroXe Nexus AI — Database Architecture

## PostgreSQL, pgvector, Apache AGE, Redis, Elasticsearch & MinIO

---

## 1. Architecture Principles

AeroXe Nexus AI follows Database-per-Microservice architecture:

| Rule | Description |
|---|---|
| Service Ownership | Each microservice owns its database |
| No Cross-DB Access | Services communicate through gRPC/NATS only |
| Event Synchronization | Data consistency through domain events |
| Read Model Optimization | Read models can be optimized separately |
| Tenant Isolation | Mandatory `tenant_id` in all business tables |

---

## 2. Storage Technology Selection

| Requirement | Technology | Purpose |
|---|---|---|
| Transaction Data | PostgreSQL 18 | Primary relational database |
| Vector Search | pgvector | Semantic search embeddings |
| Knowledge Graph | Apache AGE | Entity relationships |
| Cache | Redis | Short-term memory, sessions, rate limiting |
| Full Text Search | Elasticsearch | Document search, logs, audit |
| File Storage | MinIO | Documents, images, model files |
| Event Storage | NATS JetStream | Event streaming persistence |

---

## 3. Database-per-Service Map

```
AeroXe Nexus AI
    |
    ===================================================

    identity-service     -> identity_db    (PostgreSQL)
    ai-gateway-service   -> gateway_db     (PostgreSQL)
    agent-orchestrator   -> agent_db       (PostgreSQL)
    rag-service          -> rag_db         (PostgreSQL + pgvector)
    vision-service       -> vision_db      (PostgreSQL)
    memory-service       -> memory_db      (PostgreSQL + pgvector) + Redis
    workflow-service     -> workflow_db    (PostgreSQL)
    audit-service        -> audit_db       (PostgreSQL + Elasticsearch)

    ===================================================
```

---

## 4. Identity Database (identity_db)

### users

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    mfa_enabled BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
```

### roles

```sql
CREATE TABLE roles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### permissions

```sql
CREATE TABLE permissions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL
);
```

### user_roles

```sql
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY(user_id, role_id)
);
```

### tenants

```sql
CREATE TABLE tenants (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    status VARCHAR(50) NOT NULL DEFAULT 'pending_kyc',
    kyc_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    kyc_submitted_at TIMESTAMP,
    kyc_reviewed_at TIMESTAMP,
    kyc_reviewed_by BIGINT,
    settings JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### kyc_documents

```sql
CREATE TABLE kyc_documents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    document_type VARCHAR(100) NOT NULL,
    filename TEXT NOT NULL,
    storage_path TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'uploaded',
    reviewed_at TIMESTAMP,
    reviewed_by BIGINT,
    rejection_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_tenant ON kyc_documents(tenant_id);
```

---

## 5. Gateway Database (gateway_db)

### ai_sessions

```sql
CREATE TABLE ai_sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB
);
```

### ai_requests

```sql
CREATE TABLE ai_requests (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES ai_sessions(id),
    tenant_id BIGINT NOT NULL,
    prompt TEXT NOT NULL,
    model VARCHAR(100),
    agent VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    latency_ms FLOAT,
    tokens_used INT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_requests_session ON ai_requests(session_id);
CREATE INDEX idx_requests_tenant ON ai_requests(tenant_id, created_at DESC);
```

---

## 6. Agent Database (agent_db)

### agents

```sql
CREATE TABLE agents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    system_prompt TEXT,
    capabilities JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### agent_executions

```sql
CREATE TABLE agent_executions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    agent_id BIGINT NOT NULL REFERENCES agents(id),
    task TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    plan JSONB,
    result JSONB,
    tokens_used INT,
    latency_ms FLOAT,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_executions_tenant ON agent_executions(tenant_id, started_at DESC);
```

### agent_steps

```sql
CREATE TABLE agent_steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    execution_id BIGINT NOT NULL REFERENCES agent_executions(id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    agent_type VARCHAR(100),
    action TEXT NOT NULL,
    tool_name VARCHAR(100),
    tool_params JSONB,
    result JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);
```

---

## 7. RAG Database (rag_db) — pgvector

### documents

```sql
CREATE TABLE documents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    filename TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'uploaded',
    size_bytes BIGINT,
    storage_path TEXT,
    chunk_count INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP
);

CREATE INDEX idx_documents_tenant ON documents(tenant_id, status);
```

### document_chunks

```sql
CREATE TABLE document_chunks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    chunk_index INT NOT NULL,
    token_count INT,
    embedding vector(768),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chunks_document ON document_chunks(document_id);
CREATE INDEX idx_chunks_embedding ON document_chunks
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
```

### document_metadata

```sql
CREATE TABLE document_metadata (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    metadata JSONB NOT NULL
);
```

### document_sets

```sql
CREATE TABLE document_sets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tags JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    document_count INT DEFAULT 0,
    total_chunks INT DEFAULT 0,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_docsets_tenant ON document_sets(tenant_id, status);
```

### document_set_documents

```sql
CREATE TABLE document_set_documents (
    document_set_id BIGINT NOT NULL REFERENCES document_sets(id) ON DELETE CASCADE,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY(document_set_id, document_id)
);

CREATE INDEX idx_dsdocs_document ON document_set_documents(document_id);
```

### agent_document_sets

```sql
CREATE TABLE agent_document_sets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    document_set_id BIGINT NOT NULL REFERENCES document_sets(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    permission_level VARCHAR(50) NOT NULL DEFAULT 'read',
    bound_by BIGINT NOT NULL,
    bound_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, document_set_id)
);

CREATE INDEX idx_agent_docsets_agent ON agent_document_sets(agent_id);
CREATE INDEX idx_agent_docsets_set ON agent_document_sets(document_set_id);
CREATE INDEX idx_agent_docsets_tenant ON agent_document_sets(tenant_id);
```

### agent_databases

```sql
CREATE TABLE agent_databases (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    connection_name VARCHAR(100) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL DEFAULT 5432,
    database_name VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    password_encrypted TEXT NOT NULL,
    ssl_mode VARCHAR(20) NOT NULL DEFAULT 'require',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    last_tested_at TIMESTAMP,
    last_test_result VARCHAR(50),
    server_version VARCHAR(50),
    discovered_tables_count INT DEFAULT 0,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, database_name)
);

CREATE INDEX idx_agent_dbs_agent ON agent_databases(agent_id);
CREATE INDEX idx_agent_dbs_tenant ON agent_databases(tenant_id);
```

### agent_database_tables

```sql
CREATE TABLE agent_database_tables (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_database_id BIGINT NOT NULL REFERENCES agent_databases(id) ON DELETE CASCADE,
    agent_id BIGINT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    columns JSONB NOT NULL,
    primary_key JSONB,
    row_count_estimate INT,
    bound_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(agent_database_id, table_name)
);

CREATE INDEX idx_agent_dbtables_agent ON agent_database_tables(agent_id);
CREATE INDEX idx_agent_dbtables_db ON agent_database_tables(agent_database_id);
```

---

## 8. Vision Database (vision_db)

### images

```sql
CREATE TABLE images (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_size_bytes BIGINT,
    width INT,
    height INT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### vision_analysis

```sql
CREATE TABLE vision_analysis (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    image_id BIGINT NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    model VARCHAR(100) NOT NULL DEFAULT 'qwen3-vl:4b',
    analysis_type VARCHAR(50) NOT NULL,
    description TEXT,
    confidence FLOAT,
    detected_objects JSONB,
    metadata JSONB,
    latency_ms FLOAT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### ocr_results

```sql
CREATE TABLE ocr_results (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    image_id BIGINT NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    language VARCHAR(10) DEFAULT 'en',
    confidence FLOAT,
    regions JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 9. Memory Database (memory_db)

### memories

```sql
CREATE TABLE memories (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(768),
    memory_type VARCHAR(50) NOT NULL DEFAULT 'fact',
    importance FLOAT NOT NULL DEFAULT 0.5,
    access_count INT NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMP,
    expires_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_memories_embedding ON memories
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);
CREATE INDEX idx_memories_user ON memories(user_id, tenant_id);
```

### conversation_history

```sql
CREATE TABLE conversation_history (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    tokens_used INT,
    model VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversation_session ON conversation_history(session_id);
CREATE INDEX idx_conversation_user ON conversation_history(user_id, created_at DESC);
```

---

## 10. Workflow Database (workflow_db)

### workflows

```sql
CREATE TABLE workflows (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    definition JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### workflow_instances

```sql
CREATE TABLE workflow_instances (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES workflows(id),
    tenant_id BIGINT NOT NULL,
    initiated_by BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    context JSONB,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    error_message TEXT
);
```

### workflow_steps

```sql
CREATE TABLE workflow_steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    assignee_id BIGINT,
    input JSONB,
    output JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

---

## 11. Audit Database (audit_db)

### chat_trail (Partitioned by Month)

Complete chat trail for every conversation — every message, every tool call, every response.

```sql
CREATE TABLE chat_trail (
    id              BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id       BIGINT NOT NULL,
    session_id      VARCHAR(100) NOT NULL,
    conversation_id VARCHAR(100) NOT NULL,
    message_id      VARCHAR(100) NOT NULL,
    customer_id     BIGINT,
    user_id         BIGINT,
    role            VARCHAR(20) NOT NULL,           -- user | assistant | system | tool
    content         TEXT NOT NULL,
    content_type    VARCHAR(50) DEFAULT 'text',     -- text | image | audio | tool_call | tool_result
    model_used      VARCHAR(100),                   -- which AI model responded
    tokens_input    INTEGER,
    tokens_output   INTEGER,
    latency_ms      INTEGER,
    tool_name       VARCHAR(100),                   -- tool invoked (if any)
    tool_input      JSONB,                          -- tool call arguments
    tool_output     JSONB,                          -- tool call result
    tool_status     VARCHAR(20),                    -- success | error | blocked
    safety_flag     VARCHAR(50),                    -- none | prompt_injection | jailbreak | policy_violation
    metadata        JSONB,                          -- additional context
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Monthly partitions (auto-created by pg_partman)
CREATE TABLE chat_trail_2026_01 PARTITION OF chat_trail
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE chat_trail_2026_02 PARTITION OF chat_trail
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE chat_trail_2026_03 PARTITION OF chat_trail
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
-- ... pg_partman auto-creates future partitions

-- Indexes on partitioned table
CREATE INDEX idx_ct_tenant_session ON chat_trail(tenant_id, session_id, created_at DESC);
CREATE INDEX idx_ct_conversation ON chat_trail(conversation_id, created_at);
CREATE INDEX idx_ct_customer ON chat_trail(customer_id, created_at DESC);
CREATE INDEX idx_ct_user ON chat_trail(user_id, created_at DESC);
CREATE INDEX idx_ct_model ON chat_trail(model_used, created_at DESC);
CREATE INDEX idx_ct_safety ON chat_trail(safety_flag) WHERE safety_flag != 'none';
CREATE INDEX idx_ct_tool ON chat_trail(tool_name, created_at DESC);
```

### audit_events (Partitioned by Month)

```sql
CREATE TABLE audit_events (
    id              BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id       BIGINT NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    actor_user_id   BIGINT,
    actor_role      VARCHAR(50),
    actor_ip        INET,
    actor_user_agent TEXT,
    resource_type   VARCHAR(50),
    resource_id     BIGINT,
    action          VARCHAR(100) NOT NULL,
    result          VARCHAR(20) NOT NULL,           -- success | failure | blocked
    details         JSONB,
    trace_id        VARCHAR(100),
    request_id      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE audit_events_2026_01 PARTITION OF audit_events
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE audit_events_2026_02 PARTITION OF audit_events
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE audit_events_2026_03 PARTITION OF audit_events
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
-- ... pg_partman auto-creates future partitions

CREATE INDEX idx_audit_tenant ON audit_events(tenant_id, created_at DESC);
CREATE INDEX idx_audit_event_type ON audit_events(event_type, created_at DESC);
CREATE INDEX idx_audit_actor ON audit_events(actor_user_id, created_at DESC);
CREATE INDEX idx_audit_resource ON audit_events(resource_type, resource_id);
CREATE INDEX idx_audit_trace ON audit_events(trace_id);
CREATE INDEX idx_audit_created ON audit_events(created_at);
```

### Partition Management (pg_partman)

```sql
-- Auto-create next 3 months of partitions
SELECT partman.create_parent(
    p_parent_table := 'public.chat_trail',
    p_control := 'created_at',
    p_type := 'range',
    p_premake := 3
);

SELECT partman.create_parent(
    p_parent_table := 'public.audit_events',
    p_control := 'created_at',
    p_type := 'range',
    p_premake := 3
);

-- Retention: drop partitions older than 2 years
UPDATE partman.part_config
SET retention = '2 years',
    retention_keep_table = false
WHERE parent_table = 'public.chat_trail';

UPDATE partman.part_config
SET retention = '2 years',
    retention_keep_table = false
WHERE parent_table = 'public.audit_events';
```

### Partition Strategy

| Table | Partition Key | Interval | Hot | Warm | Cold | Archive |
|---|---|---|---|---|---|---|
| `chat_trail` | `created_at` | Monthly | 0-3 months | 3-12 months | 1-2 years | MinIO |
| `audit_events` | `created_at` | Monthly | 0-3 months | 3-12 months | 1-2 years | MinIO |

### Retention Policies

| Tier | Age | Storage | Compression | Access |
|---|---|---|---|---|
| Hot | 0-3 months | PostgreSQL (NVMe) | None | Full index scan |
| Warm | 3-12 months | PostgreSQL (SSD) | LZ4 | Indexed scan |
| Cold | 1-2 years | PostgreSQL (HDD) | ZSTD | Time-range scan |
| Archive | 2+ years | MinIO (S3) | ZSTD | Restore on demand |

### chat_trail Query Examples

```sql
-- Get all messages in a conversation
SELECT * FROM chat_trail
WHERE conversation_id = 'conv_abc123'
ORDER BY created_at ASC;

-- Get all conversations for a customer in last 7 days
SELECT DISTINCT conversation_id, MIN(created_at) AS started, MAX(created_at) AS ended
FROM chat_trail
WHERE customer_id = 12345
  AND created_at > NOW() - INTERVAL '7 days'
GROUP BY conversation_id
ORDER BY started DESC;

-- Get tool usage stats for a tenant
SELECT tool_name, COUNT(*) AS invocations,
       AVG(latency_ms) AS avg_latency,
       SUM(CASE WHEN tool_status = 'error' THEN 1 ELSE 0 END) AS errors
FROM chat_trail
WHERE tenant_id = 1
  AND tool_name IS NOT NULL
  AND created_at > NOW() - INTERVAL '30 days'
GROUP BY tool_name
ORDER BY invocations DESC;

-- Get AI model performance stats
SELECT model_used, COUNT(*) AS responses,
       AVG(tokens_output) AS avg_tokens,
       AVG(latency_ms) AS avg_latency
FROM chat_trail
WHERE role = 'assistant'
  AND model_used IS NOT NULL
  AND created_at > NOW() - INTERVAL '30 days'
GROUP BY model_used;

-- Get safety-flagged conversations
SELECT session_id, conversation_id, customer_id,
       safety_flag, content, created_at
FROM chat_trail
WHERE safety_flag != 'none'
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;

-- Export conversation as JSON for compliance
SELECT jsonb_build_object(
    'conversation_id', conversation_id,
    'messages', jsonb_agg(
        jsonb_build_object(
            'role', role,
            'content', content,
            'model', model_used,
            'tokens', tokens_input + tokens_output,
            'timestamp', created_at
        ) ORDER BY created_at
    )
) AS conversation_json
FROM chat_trail
WHERE conversation_id = 'conv_abc123'
GROUP BY conversation_id;
```

---

## 12. Elasticsearch Indices

| Index | Purpose | Retention |
|---|---|---|
| `ai_logs` | AI conversation logs | 90 days |
| `audit_events` | Audit trail (searchable) | 365 days |
| `documents` | Full-text document search | Permanent |
| `security_events` | Security alerts | 365 days |

---

## 13. MinIO Buckets

| Bucket | Purpose | Versioning |
|---|---|---|
| `aeroxe-documents` | RAG documents | Enabled |
| `aeroxe-images` | Vision analysis images | Enabled |
| `aeroxe-model-files` | Custom model files | Enabled |
| `aeroxe-backups` | System backups | Enabled |

---

## 14. Database Event Synchronization

### Pattern: Event-Driven Data Sync

```
rag-service
    |
    v  DocumentProcessed event
NATS JetStream
    |
    v
knowledge-graph-service
    |
    v  Update Apache AGE relationships
```

### Pattern: CQRS (Optional)

For high-read services:
- Write model: PostgreSQL (normalized)
- Read model: Denormalized/optimized tables
- Sync via NATS events

---

## 15. Repository Pattern (Rust)

```rust
#[async_trait]
pub trait AgentRepository: Send + Sync {
    async fn save(&self, agent: Agent) -> Result<AgentId>;
    async fn find_by_id(&self, id: AgentId) -> Result<Option<Agent>>;
    async fn find_by_tenant(&self, tenant_id: TenantId) -> Result<Vec<Agent>>;
    async fn delete(&self, id: AgentId) -> Result<()>;
}
```

---

## 16. Migration Strategy

### Tool: SeaORM Migrate (Rust)

```
migrations/
├── 001_create_users.sql
├── 002_create_roles.sql
├── 003_create_permissions.sql
├── 004_create_ai_sessions.sql
├── 005_create_documents.sql
├── 006_enable_pgvector.sql
├── 007_create_memories.sql
└── ...
```

---

## 17. Backup Strategy

| Component | Strategy | Frequency |
|---|---|---|
| PostgreSQL | Full backup + WAL archiving | Daily full, continuous WAL |
| MinIO | Versioning + Replication | Continuous |
| Redis | RDB + AOF snapshots | Every 15 min |
| NATS JetStream | Stream snapshots | Daily |
| Elasticsearch | Snapshot to MinIO | Daily |

**Recovery Targets:**
- RPO: < 15 minutes
- RTO: < 2 hours

---

## 18. Performance Targets

| Component | Target |
|---|---|
| Vector Search (pgvector) | < 200ms |
| SQL Query (PostgreSQL) | < 2s |
| Redis Lookup | < 10ms |
| PostgreSQL API Query | < 100ms |
| Elasticsearch Search | < 300ms |
| MinIO Upload | < 1s |
