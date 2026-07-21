# AeroXe Nexus AI — Database Architecture

## Shared PostgreSQL with Schema-per-Module + pgvector + Redis + MinIO

---

## 1. Architecture Principles

AeroXe Nexus AI uses a **shared PostgreSQL cluster with schema-per-module** architecture. This provides the **isolation benefits of microservices** with the **operational simplicity of a monolith**.

| Rule | Description |
|---|---|
| **Schema-per-BoundedContext** | Each module owns a PostgreSQL schema (namespace) |
| **No Cross-Schema Access via SQL** | Modules access other module's data only through Rust trait methods |
| **Schema = Future Service Boundary** | Any schema can be extracted to a standalone database service |
| **Shared Cluster** | Single PostgreSQL cluster for all modules (replication + failover) |
| **Mandatory tenant_id** | All business tables include `tenant_id` for multi-tenancy |

### Why Schema-per-Module (Not Database-per-Service)

| Aspect | Microservice DB-per-Service | Modular Monolith Schema-per-Module |
|---|---|---|
| Transactions | Distributed (2PC / Saga) | Standard ACID (same connection) |
| Schema changes | N separate migrations | Ordered migrations with module prefix |
| Query across contexts | gRPC/NATS joins | Trait method calls + in-process |
| Backup | N separate backups | Single pg_dump |
| Future extraction | N/A | Move schema to new cluster |
| Operational cost | High (N databases) | Low (1 database cluster) |

---

## 2. Storage Technology Map

| Requirement | Technology | Module Users |
|---|---|---|
| Transactional Data | **PostgreSQL 18** | All modules |
| Vector Search | pgvector (extension) | nexus-rag, nexus-memory |
| Knowledge Graph | Apache AGE (extension) | nexus-rag |
| Cache / Short-Term Memory | **Redis** | nexus-gateway, nexus-memory, nexus-identity |
| Full-Text Search | **Elasticsearch** | nexus-rag, nexus-audit |
| File Storage | **MinIO** | nexus-rag, nexus-vision |
| Event Streaming | **NATS JetStream** | All modules (async only) |

---

## 3. Schema-per-Module Map

```
PostgreSQL 18 Cluster
│
├── Schema: identity_
│   └── Module: nexus-identity
│
├── Schema: ai_
│   └── Module: nexus-ai-gateway
│
├── Schema: agent_
│   └── Module: nexus-agent
│
├── Schema: rag_
│   └── Module: nexus-rag
│
├── Schema: vision_
│   └── Module: nexus-vision
│
├── Schema: sql_
│   └── Module: nexus-sql-agent
│
├── Schema: memory_
│   └── Module: nexus-memory
│
├── Schema: workflow_
│   └── Module: nexus-workflow
│
├── Schema: security_
│   └── Module: nexus-security-ai
│
├── Schema: audit_
│   └── Module: nexus-audit
│
├── Schema: notif_
│   └── Module: nexus-notification
│
├── Schema: config_
│   └── Module: nexus-config
│
├── Schema: models_
│   └── Module: nexus-model-registry
│
└── Schema: eco_
    └── Module: nexus-ecosystem
```

---

## 4. Identity Schema (identity_)

Module: `nexus-identity`

```sql
CREATE SCHEMA IF NOT EXISTS identity;

CREATE TABLE identity.users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret TEXT,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE identity.roles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE
);

CREATE TABLE identity.permissions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL
);

CREATE TABLE identity.user_roles (
    user_id BIGINT NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES identity.roles(id) ON DELETE CASCADE,
    PRIMARY KEY(user_id, role_id)
);

CREATE TABLE identity.role_permissions (
    role_id BIGINT NOT NULL REFERENCES identity.roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES identity.permissions(id) ON DELETE CASCADE,
    PRIMARY KEY(role_id, permission_id)
);

CREATE TABLE identity.tenants (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    status VARCHAR(50) NOT NULL DEFAULT 'pending_kyc',
    kyc_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    settings JSONB
);

CREATE TABLE identity.kyc_documents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES identity.tenants(id),
    document_type VARCHAR(100) NOT NULL,
    filename TEXT NOT NULL,
    storage_path TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'uploaded'
);

CREATE TABLE identity.api_keys (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    key_hash TEXT NOT NULL,
    key_prefix VARCHAR(10) NOT NULL,
    scopes TEXT[],
    expires_at TIMESTAMP
);
```

---

## 5. AI Gateway Schema (ai_)

Module: `nexus-ai-gateway`

```sql
CREATE SCHEMA IF NOT EXISTS ai;

CREATE TABLE ai.sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB
);

CREATE TABLE ai.requests (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES ai.sessions(id),
    tenant_id BIGINT NOT NULL,
    prompt TEXT NOT NULL,
    model VARCHAR(100),
    agent VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    latency_ms FLOAT,
    tokens_used INT
);
```

---

## 6. Agent Schema (agent_)

Module: `nexus-agent`

```sql
CREATE SCHEMA IF NOT EXISTS agent;

CREATE TABLE agent.agents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    system_prompt TEXT,
    capabilities JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'active'
);

CREATE TABLE agent.executions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    agent_id BIGINT NOT NULL REFERENCES agent.agents(id),
    task TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    plan JSONB,
    result JSONB,
    tokens_used INT,
    latency_ms FLOAT
);

CREATE TABLE agent.steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    execution_id BIGINT NOT NULL REFERENCES agent.executions(id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    agent_type VARCHAR(100),
    action TEXT NOT NULL,
    tool_name VARCHAR(100),
    tool_params JSONB,
    result JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending'
);

CREATE TABLE agent.document_sets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES agent.agents(id) ON DELETE CASCADE,
    document_set_id BIGINT NOT NULL,  -- references rag.document_sets
    tenant_id BIGINT NOT NULL,
    permission_level VARCHAR(50) NOT NULL DEFAULT 'read',
    UNIQUE(agent_id, document_set_id)
);

CREATE TABLE agent.databases (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES agent.agents(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    connection_name VARCHAR(100) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL DEFAULT 5432,
    database_name VARCHAR(100) NOT NULL,
    password_encrypted TEXT NOT NULL,
    ssl_mode VARCHAR(20) NOT NULL DEFAULT 'require',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    UNIQUE(agent_id, database_name)
);

CREATE TABLE agent.database_tables (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_database_id BIGINT NOT NULL REFERENCES agent.databases(id) ON DELETE CASCADE,
    agent_id BIGINT NOT NULL REFERENCES agent.agents(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    columns JSONB NOT NULL,
    primary_key JSONB,
    UNIQUE(agent_database_id, table_name)
);
```

---

## 7. RAG Schema (rag_) — pgvector

Module: `nexus-rag`

```sql
CREATE SCHEMA IF NOT EXISTS rag;

-- Enable pgvector extension for schema
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE rag.documents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    filename TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'uploaded',
    size_bytes BIGINT,
    storage_path TEXT,
    chunk_count INT DEFAULT 0
);

CREATE TABLE rag.chunks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES rag.documents(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    chunk_index INT NOT NULL,
    token_count INT,
    embedding vector(768),
    metadata JSONB
);

CREATE INDEX idx_rag_chunks_embedding ON rag.chunks
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE TABLE rag.document_sets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tags JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    document_count INT DEFAULT 0
);

CREATE TABLE rag.document_set_documents (
    document_set_id BIGINT NOT NULL REFERENCES rag.document_sets(id) ON DELETE CASCADE,
    document_id BIGINT NOT NULL REFERENCES rag.documents(id) ON DELETE CASCADE,
    PRIMARY KEY(document_set_id, document_id)
);
```

---

## 8. Vision Schema (vision_)

Module: `nexus-vision`

```sql
CREATE SCHEMA IF NOT EXISTS vision;

CREATE TABLE vision.images (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_size_bytes BIGINT,
    width INT, height INT
);

CREATE TABLE vision.analysis (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    image_id BIGINT NOT NULL REFERENCES vision.images(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    model VARCHAR(100) NOT NULL DEFAULT 'qwen3-vl:4b',
    analysis_type VARCHAR(50) NOT NULL,
    description TEXT,
    confidence FLOAT,
    detected_objects JSONB,
    latency_ms FLOAT
);

CREATE TABLE vision.ocr_results (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    image_id BIGINT NOT NULL REFERENCES vision.images(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    language VARCHAR(10) DEFAULT 'en',
    confidence FLOAT
);
```

---

## 9. Memory Schema (memory_) + Redis

Module: `nexus-memory`

```sql
CREATE SCHEMA IF NOT EXISTS memory;

CREATE TABLE memory.memories (
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
    metadata JSONB
);

CREATE INDEX idx_memory_embedding ON memory.memories
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);

CREATE TABLE memory.conversation_history (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    tokens_used INT,
    model VARCHAR(100)
);
```

**Redis Usage:**

| Key Pattern | TTL | Purpose |
|---|---|---|
| `conversation:{user_id}:{session_id}` | 24h | Active conversation context |
| `rate:{tenant_id}:{category}` | 60s | Rate limiting counters |
| `session:{session_id}` | 24h | Session metadata |
| `memory:lock:{user_id}` | 30s | Distributed lock |

---

## 10. Workflow Schema (workflow_)

Module: `nexus-workflow`

```sql
CREATE SCHEMA IF NOT EXISTS workflow;

CREATE TABLE workflow.definitions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    definition JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE workflow.instances (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES workflow.definitions(id),
    tenant_id BIGINT NOT NULL,
    initiated_by BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    context JSONB
);

CREATE TABLE workflow.steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES workflow.instances(id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    assignee_id BIGINT,
    input JSONB,
    output JSONB
);

CREATE TABLE workflow.approvals (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    step_id BIGINT NOT NULL REFERENCES workflow.steps(id) ON DELETE CASCADE,
    approver_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    comment TEXT
);
```

---

## 11. Audit Schema (audit_) — Partitioned Tables

Module: `nexus-audit`

```sql
CREATE SCHEMA IF NOT EXISTS audit;

-- Full chat trail — every message in every conversation
CREATE TABLE audit.chat_trail (
    id              BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id       BIGINT NOT NULL,
    session_id      VARCHAR(100) NOT NULL,
    conversation_id VARCHAR(100) NOT NULL,
    message_id      VARCHAR(100) NOT NULL,
    customer_id     BIGINT,
    user_id         BIGINT,
    role            VARCHAR(20) NOT NULL,
    content         TEXT NOT NULL,
    content_type    VARCHAR(50) DEFAULT 'text',
    model_used      VARCHAR(100),
    tokens_input    INTEGER,
    tokens_output   INTEGER,
    latency_ms      INTEGER,
    tool_name       VARCHAR(100),
    tool_input      JSONB,
    tool_output     JSONB,
    tool_status     VARCHAR(20),
    safety_flag     VARCHAR(50),
    metadata        JSONB,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Audit events — all non-chat actions
CREATE TABLE audit.events (
    id              BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id       BIGINT NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    actor_user_id   BIGINT,
    actor_role      VARCHAR(50),
    actor_ip        INET,
    resource_type   VARCHAR(50),
    resource_id     BIGINT,
    action          VARCHAR(100) NOT NULL,
    result          VARCHAR(20) NOT NULL,
    details         JSONB,
    trace_id        VARCHAR(100),
    request_id      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Partition management (pg_partman auto-creates monthly partitions)
SELECT partman.create_parent(
    p_parent_table := 'audit.chat_trail',
    p_control := 'created_at',
    p_type := 'range',
    p_premake := 3
);

SELECT partman.create_parent(
    p_parent_table := 'audit.events',
    p_control := 'created_at',
    p_type := 'range',
    p_premake := 3
);
```

### Retention Policy

| Tier | Age | Storage | Compression |
|---|---|---|---|
| Hot | 0-3 months | NVMe | None |
| Warm | 3-12 months | SSD | LZ4 |
| Cold | 1-2 years | HDD | ZSTD |
| Archive | 2+ years | MinIO | ZSTD |

---

## 12. Elasticsearch Indices

| Index | Module | Purpose | Retention |
|---|---|---|---|
| `ai_logs` | nexus-audit | AI conversation logs | 90 days |
| `audit_events` | nexus-audit | Audit trail | 365 days |
| `documents` | nexus-rag | Full-text document search | Permanent |
| `security_events` | nexus-security-ai | Security alerts | 365 days |

---

## 13. MinIO Buckets

| Bucket | Module | Purpose | Versioning |
|---|---|---|---|
| `aeroxe-documents` | nexus-rag | RAG documents | Enabled |
| `aeroxe-images` | nexus-vision | Vision analysis images | Enabled |
| `aeroxe-model-files` | nexus-model-registry | Custom model files | Enabled |
| `aeroxe-backups` | All | System backups | Enabled |

---

## 14. Repository Pattern (Rust + SeaORM)

```rust
// nexus-agent/src/infrastructure/persistence/agent_repo.rs
#[async_trait]
impl AgentRepository for PostgresAgentRepository {
    async fn save(&self, agent: Agent) -> Result<AgentId> {
        // Uses SeaORM entity mapped to agent.agents table
        let model = AgentModel {
            name: agent.name,
            agent_type: agent.agent_type.to_string(),
            // ...
        };
        let result = model.insert(&self.db).await?;
        Ok(AgentId(result.last_insert_id))
    }

    async fn find_by_id(&self, id: AgentId) -> Result<Option<Agent>> {
        let model = AgentModel::find_by_id(id.0)
            .one(&self.db)
            .await?;
        Ok(model.map(Agent::from))
    }
}
```

---

## 15. Migration Strategy

### Tool: SeaORM Migrations

```
migrations/
├── 001_identity_create_schema.sql
├── 002_identity_create_users.sql
├── 003_ai_create_schema.sql
├── 004_ai_create_sessions.sql
├── 005_agent_create_schema.sql
├── 006_rag_create_schema.sql
├── 007_rag_enable_pgvector.sql
├── 008_vision_create_schema.sql
├── 009_memory_create_schema.sql
├── 010_workflow_create_schema.sql
├── 011_audit_create_schema.sql
└── ...
```

Each migration creates its module's schema first, then the tables within that schema.

---

## 16. Backup Strategy

| Component | Strategy | Frequency |
|---|---|---|
| PostgreSQL | Full + WAL archiving | Daily full, continuous WAL |
| MinIO | Versioning + Replication | Continuous |
| Redis | RDB + AOF snapshots | Every 15 min |
| NATS JetStream | Stream snapshots | Daily |
| Elasticsearch | Snapshot to MinIO | Daily |

**Recovery Targets:** RPO < 15 min, RTO < 2 hours
