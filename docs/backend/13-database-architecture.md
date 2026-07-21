# AeroXe Nexus AI — Database Architecture

## Shared PostgreSQL with Schema-per-Module + SeaORM (No Raw SQL) + pgvector + Redis + MinIO

---

## 1. Architecture Principles

AeroXe Nexus AI uses a **shared PostgreSQL cluster with schema-per-module** architecture. This provides the **isolation benefits of microservices** with the **operational simplicity of a monolith**. **No raw SQL is used anywhere** — all database access goes through SeaORM entity models.

| Rule | Description |
|---|---|
| **Schema-per-BoundedContext** | Each module owns a PostgreSQL schema (namespace) |
| **Service-Wise DB Isolation** | Each service has its own schema for easy extraction to microservices |
| **No Cross-Schema Access via SQL** | Modules access other module's data only through Rust trait methods |
| **No Raw SQL** | All DB access through SeaORM entities, models, and query builders |
| **Single ORM** | SeaORM is the only ORM — unified across all modules |
| **Schema = Future Service Boundary** | Any schema can be extracted to a standalone database service |
| **Shared Cluster** | Single PostgreSQL cluster for all modules (replication + failover) |
| **Mandatory tenant_id** | All business tables include `tenant_id` for multi-tenancy |

### Service-Wise DB Isolation

Each service/module has its own PostgreSQL schema, enabling:
- **Independent schema migrations** per service
- **Easy microservice extraction** — move schema to standalone DB
- **Clear ownership** — each team owns their schema
- **Security isolation** — cross-schema access blocked by design

| Service | Schema | Tables |
|---|---|---|
| identity | `identity` | users, roles, permissions, tenants, api_keys, sessions, kyc_documents |
| customer | `customer` | customers, addresses |
| ai-gateway | `ai` | sessions, requests |
| agent | `agent` | agents, executions, steps, document_sets, databases, database_tables |
| rag | `rag` | documents, chunks, document_metadata, document_sets |
| vision | `vision` | images, analysis, ocr_results |
| sql-agent | `sql` | (via agent schema) |
| memory | `memory` | memories, conversation_history |
| workflow | `workflow` | definitions, instances, steps, approvals |
| security | `security` | (scan results) |
| audit | `audit` | chat_trail, events (partitioned) |
| telephony | `telephony` | calls, recordings, transcripts, phone_numbers, queues, voicemails, ivr_flows, caller_auth, fraud_checks, audio_quality |
| conversation | `conversation` | conversations, messages, entities, sentiment |
| stt | `stt` | sessions, segments, models |
| tts | `tts` | voices, voice_clones, synthesis_log, post_call_surveys |
| analytics | `analytics` | conversation_metrics, call_metrics, agent_metrics, cost_tracking, snapshots |
| webhook | `webhook` | subscriptions, deliveries |
| outbound | `outbound` | campaigns, targets, callbacks, dnc_list |
| notification | `notif` | (notification templates) |
| model-registry | `models` | (model registry) |
| config | `config` | (configuration) |
| ecosystem | `eco` | (integration) |

| Aspect | Microservice DB-per-Service | Modular Monolith Schema-per-Module |
|---|---|---|
| Transactions | Distributed (2PC / Saga) | Standard ACID (same connection) |
| Schema changes | N separate migrations | Ordered SeaORM migrations with module prefix |
| Query across contexts | gRPC/NATS joins | Trait method calls + in-process |
| ORM | Each service chooses | SeaORM unified across all modules |
| Backup | N separate backups | Single pg_dump |
| Future extraction | N/A | Move schema to new cluster |
| Operational cost | High (N databases) | Low (1 database cluster) |

---

## 2. Storage Technology Map

| Requirement | Technology | Module Users |
|---|---|---|
| Transactional Data | **PostgreSQL 18** — via **SeaORM** | All modules |
| Vector Search | pgvector (extension) — via SeaORM | rag, memory |
| Knowledge Graph | Apache AGE (extension) — via SeaORM | rag |
| Cache / Short-Term Memory | **Redis** | gateway, memory, identity |
| Full-Text Search | **Elasticsearch** | rag, audit |
| File Storage | **MinIO** | rag, vision |
| Event Streaming | **NATS JetStream** | All modules (async only) |

---

## 3. Schema-per-Module Map

```
PostgreSQL 18 Cluster — All access via SeaORM entities
│
├── Schema: identity_
│   └── Module: identity (src/modules/identity/)
│
├── Schema: customer_
│   └── Module: customer (src/modules/customer/)    ← NEW
│
├── Schema: ai_
│   └── Module: ai-gateway
│
├── Schema: agent_
│   └── Module: agent
│
├── Schema: rag_
│   └── Module: rag
│
├── Schema: vision_
│   └── Module: vision
│
├── Schema: sql_
│   └── Module: sql-agent
│
├── Schema: memory_
│   └── Module: memory
│
├── Schema: workflow_
│   └── Module: workflow
│
├── Schema: security_
│   └── Module: security
│
├── Schema: audit_
│   └── Module: audit
│
├── Schema: notif_
│   └── Module: notification
│
├── Schema: config_
│   └── Module: config
│
├── Schema: models_
│   └── Module: model-registry
│
└── Schema: eco_
    └── Module: ecosystem

├── Schema: telephony_
│   └── Module: telephony (Voice channel, calls)
│
├── Schema: conversation_
│   └── Module: conversation (State machine, context)
│
├── Schema: stt_
│   └── Module: stt (Speech-to-Text)
│
├── Schema: tts_
│   └── Module: tts (Text-to-Speech)
│
├── Schema: analytics_
│   └── Module: analytics (Metrics, BI)
│
├── Schema: webhook_
│   └── Module: webhook (Event delivery)
│
└── Schema: outbound_
    └── Module: outbound (Campaigns, callbacks)

├── Schema: outbox_
│   └── Pattern: Transactional Outbox (reliable event delivery)
│
├── Schema: distributed_locks_
│   └── Pattern: Distributed Locking (Redlock)
│
├── Schema: distributed_cache_
│   └── Pattern: Distributed Caching (multi-tier)
│
└── Schema: ledger_
    └── Pattern: Double Entry Ledger (financial transactions)
```

---

## 4. SeaORM Entity Pattern (No Raw SQL)

All database access uses SeaORM entities. **No raw SQL strings anywhere in the codebase.**

### Entity Definition Example (identity module)

```rust
// src/modules/identity/infrastructure/repository/user_entity.rs
use sea_orm::entity::prelude::*;

#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "users", schema_name = "identity")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: i64,
    #[sea_orm(unique)]
    pub email: String,
    pub password_hash: String,
    pub display_name: Option<String>,
    pub status: String,
    pub mfa_enabled: bool,
    pub mfa_secret: Option<String>,
    pub last_login_at: Option<DateTime>,
    pub created_at: DateTime,
    pub updated_at: DateTime,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(has_many = "super::user_roles::Entity")]
    UserRoles,
    #[sea_orm(has_many = "super::sessions::Entity")]
    Sessions,
    #[sea_orm(has_many = "super::api_keys::Entity")]
    ApiKeys,
}

impl ActiveModelBehavior for ActiveModel {}
```

### Repository Pattern (SeaORM Queries)

```rust
// src/modules/identity/infrastructure/repository/postgres_user_repository.rs
#[async_trait]
pub trait UserRepository: Send + Sync {
    async fn find_by_id(&self, id: i64) -> Result<Option<User>, RepositoryError>;
    async fn find_by_email(&self, email: &str) -> Result<Option<User>, RepositoryError>;
    async fn find_by_tenant(&self, tenant_id: i64) -> Result<Vec<User>, RepositoryError>;
    async fn save(&self, user: &User) -> Result<User, RepositoryError>;
    async fn update(&self, user: &User) -> Result<User, RepositoryError>;
    async fn delete(&self, id: i64) -> Result<bool, RepositoryError>;
}

pub struct PostgresUserRepository {
    db: DatabaseConnection,
}

#[async_trait]
impl UserRepository for PostgresUserRepository {
    async fn find_by_email(&self, email: &str) -> Result<Option<User>, RepositoryError> {
        let model = Entity::find()
            .filter(ModelColumn::Email.eq(email))
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        Ok(model.map(User::from))
    }
    
    async fn save(&self, user: &User) -> Result<User, RepositoryError> {
        let active = ActiveModel {
            tenant_id: Set(user.tenant_id),
            email: Set(user.email.clone()),
            password_hash: Set(user.password_hash.clone()),
            display_name: Set(user.display_name.clone()),
            status: Set(user.status.clone()),
            ..Default::default()
        };
        let result = active.insert(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        Ok(User::from(result))
    }
    
    async fn find_by_tenant(&self, tenant_id: i64) -> Result<Vec<User>, RepositoryError> {
        let models = Entity::find()
            .filter(ModelColumn::TenantId.eq(tenant_id))
            .all(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        Ok(models.into_iter().map(User::from).collect())
    }
}
```

---

## 5. Customer Schema (customer_) — NEW

Module: `customer` (`src/modules/customer/`)

### SeaORM Entity (No Raw SQL)

```rust
// src/modules/customer/infrastructure/repository/customer_entity.rs
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "customers", schema_name = "customer")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: i64,
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
    pub status: String,          // active, suspended, inactive, archived
    pub kyc_status: String,      // pending, verified, rejected
    pub tags: Option<Json>,
    pub custom_fields: Option<Json>,
    pub notes: Option<Json>,
    pub created_at: DateTime,
    pub updated_at: DateTime,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(has_many = "super::addresses::Entity")]
    Addresses,
}

impl ActiveModelBehavior for ActiveModel {}
```

### Customer Addresses Entity

```rust
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "addresses", schema_name = "customer")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub customer_id: i64,
    pub address_type: String,    // billing, shipping, physical
    pub line1: String,
    pub line2: Option<String>,
    pub city: String,
    pub state: Option<String>,
    pub postal_code: String,
    pub country: String,
    pub is_default: bool,
    pub created_at: DateTime,
}
```

---

## 6. Identity Schema (identity_) — via SeaORM

Module: `identity` (`src/modules/identity/`)

All tables in the `identity` schema are accessed exclusively through SeaORM entities:

| Entity | Table | Schema |
|---|---|---|
| `UserEntity` | `identity.users` | `identity_` |
| `RoleEntity` | `identity.roles` | `identity_` |
| `PermissionEntity` | `identity.permissions` | `identity_` |
| `UserRoleEntity` | `identity.user_roles` | `identity_` |
| `RolePermissionEntity` | `identity.role_permissions` | `identity_` |
| `TenantEntity` | `identity.tenants` | `identity_` |
| `KycDocumentEntity` | `identity.kyc_documents` | `identity_` |
| `ApiKeyEntity` | `identity.api_keys` | `identity_` |
| `SessionEntity` | `identity.sessions` | `identity_` |

### Tenant Isolation (SeaORM Filter)

```rust
// All tenant-scoped queries use SeaORM's filter() — no raw WHERE clauses
let customers = customer::Entity::find()
    .filter(customer::Column::TenantId.eq(tenant_id))
    .all(&self.db)
    .await?;
```

---

## 7. AI Gateway Schema (ai_) — via SeaORM

Module: `ai-gateway` (`src/modules/ai-gateway/`)

| Entity | Table |
|---|---|
| `AiSessionEntity` | `ai.sessions` |
| `AiRequestEntity` | `ai.requests` |

---

## 8. Agent Schema (agent_) — via SeaORM

Module: `agent` (`src/modules/agent/`)

| Entity | Table |
|---|---|
| `AgentEntity` | `agent.agents` |
| `ExecutionEntity` | `agent.executions` |
| `StepEntity` | `agent.steps` |
| `DocumentSetBindingEntity` | `agent.document_sets` |
| `DatabaseConnectionEntity` | `agent.databases` |
| `DatabaseTableEntity` | `agent.database_tables` |

---

## 9. RAG Schema (rag_) — pgvector via SeaORM

Module: `rag` (`src/modules/rag/`)

### Vector Query via SeaORM

```rust
// pgvector queries still go through SeaORM's raw query or custom type
use sea_orm::JsonValue;

#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "chunks", schema_name = "rag")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub document_id: i64,
    pub content: String,
    pub chunk_index: i32,
    pub token_count: Option<i32>,
    pub embedding: Option<Vec<f32>>,  // vector(768)
    pub metadata: Option<JsonValue>,
}
```

| Entity | Table |
|---|---|
| `DocumentEntity` | `rag.documents` |
| `ChunkEntity` | `rag.chunks` |
| `DocumentSetEntity` | `rag.document_sets` |
| `DocumentSetDocumentEntity` | `rag.document_set_documents` |

---

## 10. Vision Schema (vision_) — via SeaORM

Module: `vision` (`src/modules/vision/`)

| Entity | Table |
|---|---|
| `ImageEntity` | `vision.images` |
| `AnalysisEntity` | `vision.analysis` |
| `OcrResultEntity` | `vision.ocr_results` |

---

## 11. Memory Schema (memory_) + Redis

Module: `memory` (`src/modules/memory/`)

| Entity | Table |
|---|---|
| `MemoryEntity` | `memory.memories` |
| `ConversationHistoryEntity` | `memory.conversation_history` |

**Redis Usage:**

| Key Pattern | TTL | Purpose |
|---|---|---|
| `conversation:v1:{user_id}:{session_id}` | 24h | Active conversation context |
| `rate:v1:{tenant_id}:{category}` | 60s | Rate limiting counters |
| `session:v1:{session_id}` | 24h | Session metadata |
| `memory:lock:v1:{user_id}` | 30s | Distributed lock |

---

## 12. Workflow Schema (workflow_) — via SeaORM

Module: `workflow` (`src/modules/workflow/`)

| Entity | Table |
|---|---|
| `WorkflowDefinitionEntity` | `workflow.definitions` |
| `WorkflowInstanceEntity` | `workflow.instances` |
| `WorkflowStepEntity` | `workflow.steps` |
| `WorkflowApprovalEntity` | `workflow.approvals` |

---

## 13. Audit Schema (audit_) — Partitioned via SeaORM

Module: `audit` (`src/modules/audit/`)

| Entity | Table |
|---|---|
| `ChatTrailEntity` | `audit.chat_trail` (partitioned) |
| `AuditEventEntity` | `audit.events` (partitioned) |

### Retention Policy

| Tier | Age | Storage | Compression |
|---|---|---|---|
| Hot | 0-3 months | NVMe | None |
| Warm | 3-12 months | SSD | LZ4 |
| Cold | 1-2 years | HDD | ZSTD |
| Archive | 2+ years | MinIO | ZSTD |

---

## 14. Elasticsearch Indices

| Index | Module | Purpose | Retention |
|---|---|---|---|
| `ai_logs_v1` | audit | AI conversation logs | 90 days |
| `audit_events_v1` | audit | Audit trail | 365 days |
| `documents_v1` | rag | Full-text document search | Permanent |
| `security_events_v1` | security | Security alerts | 365 days |

---

## 15. MinIO Buckets

| Bucket | Module | Purpose | Versioning |
|---|---|---|---|
| `aeroxe-documents` | rag | RAG documents | Enabled |
| `aeroxe-images` | vision | Vision analysis images | Enabled |
| `aeroxe-model-files` | model-registry | Custom model files | Enabled |
| `aeroxe-backups` | All | System backups | Enabled |

---

## 16. Migration Strategy

### Tool: SeaORM Migrations (Rust, not SQL)

```
Each module has its own migration directory with versioned Rust files:

src/modules/identity/migrations/
├── mod.rs
├── m20250701_000001_create_schema.rs
├── m20250701_000002_create_users_table.rs
├── m20250701_000003_create_roles_table.rs
├── m20250701_000004_create_permissions_table.rs
├── m20250701_000005_create_user_roles_table.rs
├── m20250701_000006_create_tenants_table.rs
├── m20250701_000007_create_api_keys_table.rs
├── m20250701_000008_create_sessions_table.rs
└── m20250701_000009_create_kyc_documents_table.rs

src/modules/customer/migrations/
├── mod.rs
├── m20250701_000001_create_customer_schema.rs
├── m20250701_000002_create_customers_table.rs
└── m20250701_000003_create_addresses_table.rs

src/modules/ai-gateway/migrations/
src/modules/agent/migrations/
src/modules/rag/migrations/
...
```

Example migration:

```rust
use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager.create_schema(
            Schema::create("customer".to_string()).to_owned()
        ).await?;
        
        manager.create_table(
            Table::create()
                .table(Customers::Table)
                .if_not_exists()
                .col(ColumnDef::new(Customers::Id).big_integer().auto_increment().primary_key())
                .col(ColumnDef::new(Customers::TenantId).big_integer().not_null())
                .col(ColumnDef::new(Customers::Name).string().not_null())
                .col(ColumnDef::new(Customers::Email).string().not_null())
                .col(ColumnDef::new(Customers::Phone).string())
                .col(ColumnDef::new(Customers::Status).string().not_null().default("active"))
                .col(ColumnDef::new(Customers::CreatedAt).timestamp().not_null())
                .col(ColumnDef::new(Customers::UpdatedAt).timestamp().not_null())
                .to_owned()
        ).await?;
        Ok(())
    }
    
    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager.drop_table(Table::drop().table(Customers::Table).to_owned()).await?;
        Ok(())
    }
}
```

---

## 17. No Raw SQL Enforcement

The following rules are enforced across the entire codebase:

| Rule | Enforcement |
|---|---|
| No SQL strings in code | All DB access through SeaORM entity models |
| No string interpolation | Use SeaORM's `filter()`, `find()`, `insert()`, `update()` |
| No raw queries | SeaORM's `QueryBuilder` for complex queries |
| Migrations in Rust | SeaORM migration framework (not SQL files) |
| Schema creation | `manager.create_schema()` in migrations |
| Vector queries | SeaORM with custom pgvector type support |

---

## 18. Backup Strategy

| Component | Strategy | Frequency |
|---|---|---|
| PostgreSQL | Full + WAL archiving | Daily full, continuous WAL |
| MinIO | Versioning + Replication | Continuous |
| Redis | RDB + AOF snapshots | Every 15 min |
| NATS JetStream | Stream snapshots | Daily |
| Elasticsearch | Snapshot to MinIO | Daily |

**Recovery Targets:** RPO < 15 min, RTO < 2 hours
