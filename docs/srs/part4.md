# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 4 — Database Architecture & Data Design

## Shared PostgreSQL + Schema-per-Module + SeaORM (No Raw SQL) + pgvector + Apache AGE + Redis + Elasticsearch + MinIO

---

# 1. Database Architecture Principles

AeroXe Nexus AI follows **Schema-per-Module** architecture — all modules share a single PostgreSQL cluster.

Rules:

✅ All modules share a single PostgreSQL cluster
✅ Each module owns a PostgreSQL schema (namespace)
✅ **No raw SQL** — all access through SeaORM entities and models
✅ No direct cross-schema SQL access — communicate through Rust trait methods
✅ Data consistency through events (NATS JetStream)
✅ Schema = Future service boundary (extractable later)
✅ Tenant isolation mandatory (`tenant_id` column in all business tables)

---

# 2. Data Architecture Overview

```text
                         AeroXe Nexus AI


                               |

                 Schema-per-Module Ownership


                               |


================================================================

Single PostgreSQL 16 Cluster


 Identity Module (schema: identity_)

        → SeaORM entities


 Customer Module (schema: customer_)  ← NEW

        → SeaORM entities


 Agent Module (schema: agent_)

        → SeaORM entities


 AI Gateway Module (schema: ai_)

        → SeaORM entities


 RAG Module (schema: rag_) + pgvector

        → SeaORM entities + vector columns


 Vision Module (schema: vision_)

        → SeaORM entities


 Memory Module (schema: memory_) + Redis

        → SeaORM entities + Redis keys


 Workflow Module (schema: workflow_)

        → SeaORM entities


 Audit Module (schema: audit_) + Elasticsearch

        → SeaORM entities (partitioned) + ES indices


================================================================

```

---

# 3. Storage Technology Selection

| Requirement      | Technology     |
| ---------------- | -------------- |
| Transaction Data | PostgreSQL 16  |
| Vector Search    | pgvector       |
| Knowledge Graph  | Apache AGE     |
| Cache            | Redis          |
| Full Text Search | Elasticsearch  |
| File Storage     | MinIO          |
| Event Storage    | NATS JetStream |

---

# 4. Multi-Tenant Data Architecture

AeroXe Nexus AI supports:

* Multiple companies
* Multiple business units
* Multiple users
* Multiple AI agents

Every business table must include:

```rust
// SeaORM Entity - Every business table includes tenant_id
tenant_id: i64,
```

Example:```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "ai_sessions")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: i64,
    pub user_id: i64,
    pub created_at: chrono::NaiveDateTime,
}
```

---

# 5. Identity Module Database

Database:

```text
identity_db
```

Purpose:

Authentication and authorization.

---

# 5.1 Users Table```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "users")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: i64,
    pub email: Option<String>,
    pub password_hash: Option<String>,
    pub status: Option<String>,
    pub created_at: Option<chrono::NaiveDateTime>,
    pub updated_at: Option<chrono::NaiveDateTime>,
}
```

---

# 5.2 Roles```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "roles")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: Option<i64>,
    pub name: Option<String>,
    pub description: Option<String>,
}
```

---

# 5.3 Permissions```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "permissions")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub name: Option<String>,
    pub resource: Option<String>,
    pub action: Option<String>,
}
```

---

# 5.4 User Roles```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "user_roles")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub user_id: i64,
    #[sea_orm(primary_key)]
    pub role_id: i64,
}
```

---

# 5.5 Tenants```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "tenants")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub name: String,
    pub slug: String,
    pub plan: String,
    pub status: String,
    pub kyc_status: String,
    pub kyc_submitted_at: Option<chrono::NaiveDateTime>,
    pub kyc_reviewed_at: Option<chrono::NaiveDateTime>,
    pub kyc_reviewed_by: Option<i64>,
    pub settings: Option<serde_json::Value>,
    pub created_at: chrono::NaiveDateTime,
}
```

---

# 5.6 KYC Documents```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "kyc_documents")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: i64,
    pub document_type: String,
    pub filename: String,
    pub storage_path: String,
    pub status: String,
    pub reviewed_at: Option<chrono::NaiveDateTime>,
    pub reviewed_by: Option<i64>,
    pub rejection_reason: Option<String>,
    pub created_at: chrono::NaiveDateTime,
}
```

---

# 5.7 Document Sets```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "document_sets")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub tags: Option<serde_json::Value>,
    pub status: String,
    pub document_count: Option<i32>,
    pub total_chunks: Option<i32>,
    pub created_by: i64,
    pub created_at: chrono::NaiveDateTime,
    pub updated_at: chrono::NaiveDateTime,
}
```

---

# 5.8 Document Set Documents```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "document_set_documents")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub document_set_id: i64,
    #[sea_orm(primary_key)]
    pub document_id: i64,
    pub added_at: chrono::NaiveDateTime,
}
```

---

# 5.9 Agent Document Sets```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "agent_document_sets")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub agent_id: i64,
    pub document_set_id: i64,
    pub tenant_id: i64,
    pub permission_level: String,
    pub bound_by: i64,
    pub bound_at: chrono::NaiveDateTime,
}
```

---

# 5.10 Agent Databases```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "agent_databases")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub agent_id: i64,
    pub tenant_id: i64,
    pub connection_name: String,
    pub host: String,
    pub port: i32,
    pub database_name: String,
    pub username: String,
    pub password_encrypted: String,
    pub ssl_mode: String,
    pub status: String,
    pub last_tested_at: Option<chrono::NaiveDateTime>,
    pub last_test_result: Option<String>,
    pub server_version: Option<String>,
    pub discovered_tables_count: Option<i32>,
    pub created_by: i64,
    pub created_at: chrono::NaiveDateTime,
    pub updated_at: chrono::NaiveDateTime,
}
```

---

# 5.11 Agent Database Tables```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "agent_database_tables")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub agent_database_id: i64,
    pub agent_id: i64,
    pub table_name: String,
    pub columns: serde_json::Value,
    pub primary_key: Option<serde_json::Value>,
    pub row_count_estimate: Option<i32>,
    pub bound_at: chrono::NaiveDateTime,
}
```

---

# 6. AI Gateway Database

Database:

```text
gateway_db
```

Purpose:

Store AI sessions and requests.

---

# 6.1 AI Session```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "ai_sessions")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: Option<i64>,
    pub user_id: Option<i64>,
    pub started_at: Option<chrono::NaiveDateTime>,
    pub status: Option<String>,
}
```

---

# 6.2 AI Requests```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "ai_requests")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub session_id: Option<i64>,
    pub prompt: Option<String>,
    pub model: Option<String>,
    pub status: Option<String>,
    pub created_at: Option<chrono::NaiveDateTime>,
}
```

---

# 7. Agent Orchestrator Database

Database:

```text
agent_db
```

Purpose:

Track AI agent execution.

---

# 7.1 Agents```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "agents")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: i64,
    pub name: Option<String>,
    pub type_: Option<String>,  // 'type' is a Rust keyword
    pub model: Option<String>,
    pub status: Option<String>,
}
```

---

# 7.2 Agent Executions```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "agent_executions")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: Option<i64>,
    pub agent_id: Option<i64>,
    pub task: Option<String>,
    pub status: Option<String>,
    pub started_at: Option<chrono::NaiveDateTime>,
    pub completed_at: Option<chrono::NaiveDateTime>,
}
```

---

# 7.3 Agent Steps```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "agent_steps")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub execution_id: Option<i64>,
    pub step_number: Option<i32>,
    pub action: Option<String>,
    pub result: Option<serde_json::Value>,
}
```

---

# 8. RAG Database Architecture

Database:

```text
rag_db
```

Technology:

```text
PostgreSQL

+

pgvector

```

---

# 8.1 Documents Table```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "documents")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: Option<i64>,
    pub filename: Option<String>,
    pub type_: Option<String>,  // 'type' is a Rust keyword
    pub status: Option<String>,
    pub created_at: Option<chrono::NaiveDateTime>,
}
```

---

# 8.2 Document Chunks```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "document_chunks")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub document_id: Option<i64>,
    pub content: Option<String>,
    pub chunk_index: Option<i32>,
    pub embedding: Option<Vec<f32>>,  // pgvector vector(768) — nomic-embed-text via Ollama
}
```

---

# 8.3 Vector Index```rust
// SeaORM Migration - Index creation
use sea_orm_migration::prelude;

#[derive(Iden)]
pub struct DocumentChunks;

pub struct CreateEmbeddingIndex;

#[async_trait::async_trait]
impl MigrationTrait for CreateEmbeddingIndex {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_index(
                Index::create()
                    .name("embedding_index")
                    .table(DocumentChunks::Table)
                    .col(DocumentChunks::Embedding)
                    .index_type(IndexType::Custom("ivfflat".into()))
                    .using("vector_cosine_ops")
                    .to_owned(),
            )
            .await
    }
}
```

---

# 8.4 Metadata```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "document_metadata")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub document_id: Option<i64>,
    pub metadata: Option<serde_json::Value>,
}
```

Example:

```json
{
 "department":"billing",
 "category":"invoice",
 "security":"private"
}

```

---

# 9. RAG Data Flow

```text
Document Upload


      |

      |

MinIO Storage


      |

      |

Parser


      |

      |

Chunk Generator


      |

      |

Embedding Model

(nomic-embed-text: 768 dimensions, via Ollama)


      |

      |

pgvector


      |

      |

Command-R 7B


      |

      |

Answer

```

---

# 10. Knowledge Graph Database

Technology:

```text
Apache AGE
```

Purpose:

Relationship intelligence.

---

# Example Knowledge Graph

```text
Customer

 |

has

 |

Subscription

 |

connected_to

 |

ONU Device

 |

belongs_to

 |

OLT

 |

located_at

 |

City

```

---

# 10.1 Graph Entities

Nodes:

```text
Customer

Company

Device

Network

Document

Agent

```

---

# Relationships:

```text
OWNS

CONNECTED_TO

DEPENDS_ON

RELATED_TO

BELONGS_TO

```

---

# 11. Vision Module Database

Database:

```text
vision_db
```

---

# 11.1 Images```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "images")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: Option<i64>,
    pub storage_path: Option<String>,
    pub type_: Option<String>,  // 'type' is a Rust keyword
    pub created_at: Option<chrono::NaiveDateTime>,
}
```

---

# 11.2 Vision Analysis```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "vision_analysis")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub image_id: Option<i64>,
    pub model: Option<String>,
    pub description: Option<String>,
    pub confidence: Option<f64>,
    pub metadata: Option<serde_json::Value>,
}
```

---

# 11.3 OCR Results```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "ocr_results")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub image_id: Option<i64>,
    pub text: Option<String>,
}
```

---

# 12. Memory Module Database

Database:

```text
memory_db
```

---

# 12.1 Short Term Memory

Technology:

```text
Redis
```

Example:

```
conversation:{user_id}

```

Stores:

* Current conversation
* Temporary context
* Active tasks

---

# 12.2 Long Term Memory

PostgreSQL:```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "memories")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub user_id: Option<i64>,
    pub content: Option<String>,
    pub embedding: Option<Vec<f32>>,  // pgvector vector(768) — nomic-embed-text via Ollama
    pub importance: Option<f64>,
    pub created_at: Option<chrono::NaiveDateTime>,
}
```

---

# 13. Workflow Database

Database:

```text
workflow_db
```

---

# 13.1 Workflow Definition```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "workflows")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub name: Option<String>,
    pub definition: Option<serde_json::Value>,
}
```

---

# 13.2 Workflow Execution```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "workflow_instances")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub workflow_id: Option<i64>,
    pub status: Option<String>,
    pub started_at: Option<chrono::NaiveDateTime>,
}
```

---

# 14. Audit Database

Database:

```text
audit_db
```

Purpose:

Complete compliance tracking.

---

# 14.1 Audit Events```rust
// SeaORM Entity Definition
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "audit_events")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: Option<i64>,
    pub service: Option<String>,
    pub event_type: Option<String>,
    pub payload: Option<serde_json::Value>,
    pub created_at: Option<chrono::NaiveDateTime>,
}
```

---

# 15. Elasticsearch Usage

Used for:

* Logs
* AI conversations
* Audit search
* Knowledge search

Indexes:

```
ai_logs

audit_events

documents

security_events

```

---

# 16. MinIO Storage Design

Purpose:

Object storage.

Buckets:

```
aeroxe-documents

aeroxe-images

aeroxe-model-files

aeroxe-backups

```

---

# 17. Database Event Synchronization (Versioned)

Example:

Document Processing:

```text
rag module


aeroxe.v1.rag.document.processed


        |

        |

NATS JetStream


        |

        |

agent module (via subscriber)


Update Knowledge Available


```

---

# 18. Repository Pattern (SeaORM)

DDD Repository Example — **No raw SQL**:

```rust
use sea_orm::{DatabaseConnection, EntityTrait, Set, ColumnTrait, QueryFilter};
use crate::domain::entities::agent;

#[async_trait]
pub trait AgentRepository: Send + Sync {
    async fn save(&self, agent: Agent) -> Result<AgentId, RepositoryError>;
    async fn find_by_id(&self, id: i64) -> Result<Option<Agent>, RepositoryError>;
    async fn find_by_tenant(&self, tenant_id: i64) -> Result<Vec<Agent>, RepositoryError>;
}

pub struct PostgresAgentRepository {
    db: DatabaseConnection,
}

#[async_trait]
impl AgentRepository for PostgresAgentRepository {
    async fn save(&self, agent: Agent) -> Result<AgentId, RepositoryError> {
        let active = agent::ActiveModel {
            name: Set(agent.name),
            agent_type: Set(agent.agent_type.to_string()),
            model: Set(agent.model),
            status: Set(agent.status.to_string()),
            ..Default::default()
        };
        let result = active.insert(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        Ok(AgentId(result.id))
    }

    async fn find_by_id(&self, id: i64) -> Result<Option<Agent>, RepositoryError> {
        let model = agent::Entity::find_by_id(id)
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        Ok(model.map(Agent::from))
    }

    async fn find_by_tenant(&self, tenant_id: i64) -> Result<Vec<Agent>, RepositoryError> {
        let models = agent::Entity::find()
            .filter(agent::Column::TenantId.eq(tenant_id))
            .all(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        Ok(models.into_iter().map(Agent::from).collect())
    }
}
```

---

# 18.1 SeaORM Dependencies

Required `Cargo.toml` dependencies:

```toml
[dependencies]
sea-orm = { version = "1.0", features = ["sqlx-postgres", "runtime-tokio-rustls", "macros"] }
sea-orm-migration = "1.0"
sea-orm-cli = "1.0"  # Optional: for CLI migrations
chrono = { version = "0.4", features = ["serde"] }
serde_json = "1.0"
async-trait = "0.1"

[dev-dependencies]
sea-orm = { version = "1.0", features = ["sqlx-postgres", "runtime-tokio-rustls", "macros", "mock"] }
```

---

# 18.2 Entity Relationships

SeaORM entities require `Relation` enum and `Related` trait for foreign keys:

```rust
use sea_orm::entity::prelude::*;

#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "agent_executions")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: Option<i64>,
    pub agent_id: Option<i64>,
    pub task: Option<String>,
    pub status: Option<String>,
    pub started_at: Option<chrono::NaiveDateTime>,
    pub completed_at: Option<chrono::NaiveDateTime>,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(
        belongs_to = "super::agent::Entity",
        from = "Column::AgentId",
        to = "super::agent::Column::Id"
    )]
    Agent,
}

impl Related<super::agent::Entity> for Entity {
    fn to() -> RelationDef {
        Relation::Agent.def()
    }
}

impl ActiveModelBehavior for ActiveModel {}
```

---

# 19. Database Migration Strategy

Technology:

```text
SeaORM Migrate (Rust)
```

Structure:

```text
migrations/

├── src/
│
│   ├── lib.rs
│
│   ├── m20220101_000001_create_users_table.rs
│
│   ├── m20220101_000002_create_roles_table.rs
│
│   └── m20220101_000003_create_permissions_table.rs
│
└── Cargo.toml
```

Example Migration:

```rust
use sea_orm_migration::prelude;

pub struct CreateUsersTable;

#[async_trait::async_trait]
impl MigrationTrait for CreateUsersTable {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_table(
                Table::create()
                    .table(Users::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Users::Id)
                            .big_integer()
                            .not_null()
                            .auto_increment()
                            .primary_key(),
                    )
                    .col(ColumnDef::new(Users::TenantId).big_integer().not_null())
                    .col(ColumnDef::new(Users::Email).string_len(255).unique_key())
                    .col(ColumnDef::new(Users::PasswordHash).text())
                    .col(ColumnDef::new(Users::Status).string_len(50))
                    .col(ColumnDef::new(Users::Timestamp).timestamp())
                    .to_owned(),
            )
            .await
    }
}
```

---

# 20. Backup Strategy

## PostgreSQL

* Daily full backup
* WAL archiving
* Point-in-time recovery

## MinIO

* Versioning enabled
* Replication

## NATS JetStream

* Snapshot backup

---

# 21. Performance Requirements

| Component            | Target |
| -------------------- | ------ |
| Vector Search        | <200ms |
| SQL Query            | <2s    |
| Redis Lookup         | <10ms  |
| PostgreSQL API Query | <100ms |
| Elasticsearch Search | <300ms |

---

# 22. Final Database Architecture

```text
                     AeroXe Nexus AI


                            |

==================================================


PostgreSQL

├── identity_db

├── gateway_db

├── agent_db

├── rag_db

├── vision_db

├── memory_db

├── workflow_db

└── audit_db



├── telephony_db        ← NEW (Voice/Calling)

├── conversation_db     ← NEW (State Machine)

├── stt_db              ← NEW (Speech-to-Text)

├── tts_db              ← NEW (Text-to-Speech)

├── analytics_db        ← NEW (Metrics/BI)

├── webhook_db          ← NEW (Event Delivery)

├── outbound_db         ← NEW (Campaigns)

├── billing_db          ← NEW (Subscriptions)

├── outbox_db           ← NEW (Transactional Outbox)

├── distributed_locks_db ← NEW (Distributed Locking)

├── distributed_cache_db ← NEW (Distributed Caching)

├── ledger_db           ← NEW (Double Entry Ledger)



pgvector

└── Semantic Search



Apache AGE

└── Knowledge Graph



Redis

└── Memory Cache



Elasticsearch

└── Search + Analytics



MinIO

└── Object Storage


==================================================

```

---

# Part 4 Completed

Covered:

✅ Schema-per-Module architecture
✅ PostgreSQL schema design
✅ pgvector RAG database
✅ Knowledge Graph model
✅ Memory architecture
✅ Multi-tenancy
✅ Repository pattern
✅ Backup strategy
✅ Telephony schema (NEW)
✅ Conversation schema (NEW)
✅ STT/TTS schemas (NEW)
✅ Analytics schema (NEW)
✅ Webhook schema (NEW)
✅ Outbound schema (NEW)
✅ Billing schema (NEW)
✅ Outbox schema (NEW)
✅ Distributed Locking schema (NEW)
✅ Distributed Caching schema (NEW)
✅ Ledger schema (NEW)
