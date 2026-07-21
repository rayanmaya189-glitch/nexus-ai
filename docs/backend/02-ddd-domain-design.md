# AeroXe Nexus AI — DDD Domain Design

## Domain-Driven Design + Modular Monolith + Aggregate Design

---

## 1. DDD Architecture Overview

AeroXe Nexus AI is designed using **Domain-Driven Design** principles organized as a **modular monolith**. The system is divided into independent **Bounded Contexts**, each owning its business logic, database schema, and trait interfaces. Modules communicate through Rust trait methods (synchronous) and NATS events (asynchronous).

### Key Difference from Microservices

| Aspect | Microservice Architecture | Modular Monolith (This Project) |
|---|---|---|
| Communication | gRPC over network | Rust trait method calls (in-process) |
| Database | Separate DB per service | Shared PostgreSQL, schema per module |
| Deployment | N containers | 1 binary |
| Testing | Service-level integration tests | Module-level + full binary tests |
| Latency | 2-5ms per gRPC call | < 1μs per trait dispatch |
| Extractability | N/A | Any module can be extracted to a microservice later |

---

## 2. Core Domain Classification

| Domain | Type | Module Name | Schema Prefix |
|---|---|---|---|
| Agent Orchestration | **Core Domain** | `nexus-agent` | `agent_` |
| AI Gateway | **Core Domain** | `nexus-ai-gateway` | `ai_` |
| RAG Intelligence | **Core Domain** | `nexus-rag` | `rag_` |
| Vision Intelligence | **Core Domain** | `nexus-vision` | `vision_` |
| SQL Intelligence | **Core Domain** | `nexus-sql-agent` | `sql_` |
| Security Intelligence | **Core Domain** | `nexus-security-ai` | `security_` |
| Identity | Supporting | `nexus-identity` | `identity_` |
| Memory | Supporting | `nexus-memory` | `memory_` |
| Audit | Supporting | `nexus-audit` | `audit_` |
| Workflow | Supporting | `nexus-workflow` | `workflow_` |

### Infrastructure Modules

| Module | Purpose | Schema Prefix |
|---|---|---|
| `nexus-gateway` | API Gateway (axum HTTP/WS) | — (stateless) |
| `nexus-model-registry` | Ollama model management | `models_` |
| `nexus-notification` | Email, WhatsApp, push | `notif_` |
| `nexus-config` | Dynamic configuration | `config_` |
| `nexus-ecosystem` | AeroXe product connectors | `eco_` |

---

## 3. Module Internal Code Structure (DDD Layers)

Every module follows a strict layered architecture:

```
nexus-<name>/                      # Cargo crate
├── Cargo.toml
├── src/
│   ├── lib.rs                     # Public API: trait definitions + re-exports
│   ├── domain/                    # 🟢 DDD Layer: Business logic
│   │   ├── mod.rs
│   │   ├── aggregates/            # Aggregate roots with invariants
│   │   ├── entities/              # Mutable domain objects (with IDs)
│   │   ├── value_objects/         # Immutable validated types (no IDs)
│   │   └── events/                # Domain event structs + serialization
│   ├── application/               # 🟡 DDD Layer: Use cases / orchestration
│   │   ├── mod.rs
│   │   ├── commands/              # Command structs (what you tell the system to do)
│   │   ├── queries/               # Query structs (what you ask the system)
│   │   ├── handlers/              # Command/Query handler implementations
│   │   └── services/              # Application services (orchestration)
│   ├── infrastructure/            # 🔵 DDD Layer: Technical concerns
│   │   ├── mod.rs
│   │   ├── persistence/           # SeaORM repositories + migrations
│   │   ├── ollama/                # Ollama HTTP client adapter
│   │   └── nats/                  # NATS publisher/subscriber adapter
│   └── interfaces/                # 🟣 Ports & Adapters
│       ├── mod.rs
│       ├── api.rs                 # Public trait definitions (other modules consume)
│       └── events.rs              # NATS event subscribers
├── tests/
│   ├── unit/                      # 🔴 TDD: Domain unit tests (no infra)
│   ├── integration/               # 🟠 TDD: Integration tests (DB + NATS)
│   ├── contract/                  # 🟢 TDD: Module boundary contract tests
│   └── e2e/                       # 🔵 TDD: Full flow tests (all modules)
├── migrations/                    # SeaORM migration files
│   ├── 001_create_tables.sql
│   └── ...
└── Cargo.lock
```

---

## 4. Module Public API Trait Pattern

Each module exposes its public API as Rust traits. Other modules depend on the trait, not the implementation.

```rust
// nexus-rag/src/interfaces/api.rs
#[async_trait]
pub trait RagService: Send + Sync {
    async fn search(&self, query: SearchQuery) -> Result<SearchResults, RagError>;
    async fn upload_document(&self, request: UploadRequest) -> Result<DocumentStatus>;
    async fn get_document_status(&self, id: DocumentId) -> Result<Option<DocumentStatus>>;
    async fn delete_document(&self, id: DocumentId) -> Result<()>;
}

// nexus-agent uses RagService via trait, not directly
pub struct AgentServiceImpl {
    rag: Arc<dyn RagService>,     // <-- injected at binary composition
    memory: Arc<dyn MemoryService>,
    ollama: OllamaClient,
    // ...
}
```

The binary's `main.rs` wires all module implementations together:

```rust
// src/main.rs (binary crate)
#[tokio::main]
async fn main() {
    // Init infrastructure
    let db = init_db().await;
    let redis = init_redis().await;
    let nats = init_nats().await;
    let ollama = OllamaClient::new(config.ollama_url);

    // Init module implementations (each implements the trait)
    let identity = Arc::new(nexus_identity::IdentityServiceImpl::new(db.clone(), redis.clone())) as Arc<dyn IdentityService>;
    let memory = Arc::new(nexus_memory::MemoryServiceImpl::new(db.clone(), redis.clone(), ollama.clone())) as Arc<dyn MemoryService>;
    let rag = Arc::new(nexus_rag::RagServiceImpl::new(db.clone(), nats.clone(), ollama.clone())) as Arc<dyn RagService>;
    let agent = Arc::new(nexus_agent::AgentServiceImpl::new(rag.clone(), memory.clone(), ollama.clone())) as Arc<dyn AgentService>;
    // ... etc

    // Build API gateway with all module traits
    let app = nexus_gateway::build_router(AppState {
        identity, agent, rag, memory, /* ... */
    });

    // Start server
    let listener = tokio::net::TcpListener::bind("0.0.0.0:8080").await?;
    axum::serve(listener, app).await?;
}
```

---

## 5. Identity Bounded Context

**Module:** `nexus-identity`
**Schema:** `identity_` prefix in shared PostgreSQL
**Purpose:** Manage users, tenants, roles, permissions

### Aggregate: User

```
User (Aggregate Root)
├── Profile
│   ├── UserId
│   ├── EmailAddress (value object)
│   ├── DisplayName
│   └── Avatar
├── Authentication
│   ├── PasswordHash (bcrypt, cost 12)
│   ├── OTP Secret
│   └── MFA Settings
├── Roles[] (Entities)
│   ├── RoleId
│   ├── Name
│   └── Permissions[] (value objects)
└── TenantMembership
    ├── TenantId
    └── TenantRole
```

### Entities

| Entity | Attributes |
|---|---|
| User | UserId, TenantId, Email, Status, CreatedAt |
| Role | RoleId, Name, Permissions[] (shared with tenant) |
| Tenant | TenantId, Name, Plan, Settings, KYC status |
| APIKey | KeyId, TenantId, UserId, KeyHash, Scopes[] |

### Value Objects

| Value Object | Rust Type | Validation |
|---|---|---|
| `UserId` | `i64` | Positive |
| `TenantId` | `i64` | Positive |
| `EmailAddress` | `String` | Regex validated |
| `Permission` | `String` | Format: `{resource}.{action}` |
| `PasswordHash` | `String` | bcrypt verified |
| `JWTToken` | `String` | RS256 signed + exp check |

### Domain Events (NATS Subjects)

| Event | Trigger |
|---|---|
| `aeroxe.identity.user.created` | New user registration |
| `aeroxe.identity.user.updated` | Profile change |
| `aeroxe.identity.role.assigned` | Role assignment |
| `aeroxe.identity.permission.changed` | Permission modification |

### Public API Trait

```rust
#[async_trait]
pub trait IdentityService: Send + Sync {
    async fn authenticate(&self, req: AuthRequest) -> Result<AuthResponse, IdentityError>;
    async fn verify_token(&self, token: &str) -> Result<JWTClaims, IdentityError>;
    async fn check_permission(&self, req: PermissionRequest) -> Result<bool, IdentityError>;
    async fn validate_tenant(&self, tenant_id: TenantId) -> Result<Tenant, IdentityError>;
    async fn create_user(&self, req: CreateUserRequest) -> Result<User, IdentityError>;
    async fn get_user(&self, id: UserId) -> Result<Option<User>, IdentityError>;
}
```

### TDD: Command / Query Pattern

| Command | Query |
|---|---|
| `CreateUserCommand` | `GetUserQuery` |
| `AssignRoleCommand` | `GetPermissionsQuery` |
| `UpdatePermissionCommand` | `GetTenantUsersQuery` |
| `LoginCommand` | `GetTenantQuery` |

**Unit Test Example:**

```rust
#[tokio::test]
async fn test_user_creation_requires_valid_email() {
    let mut repo = MockUserRepository::new();
    let service = IdentityServiceImpl::new(repo, /* ... */);

    let result = service.create_user(CreateUserRequest {
        email: "invalid".to_string(), // no @ symbol
        password: "ValidPass123!".to_string(),
        tenant_id: TenantId(1),
    }).await;

    assert!(matches!(result, Err(IdentityError::InvalidEmail)));
}
```

---

## 6. AI Gateway Bounded Context

**Module:** `nexus-ai-gateway`
**Schema:** `ai_` prefix
**Purpose:** Central AI request lifecycle, prompt safety, model routing

### Aggregate: AIRequest

```
AIRequest (Aggregate Root)
├── RequestContext
│   ├── RequestId (UUID)
│   ├── TenantId
│   ├── UserId
│   └── TraceId
├── SecurityContext
│   ├── JWT Claims
│   ├── Permissions
│   └── Tenant Scope
└── ExecutionPlan
    ├── TargetAgent (via nexus-agent)
    ├── ModelPreference
    └── Priority
```

### Public API Trait

```rust
#[async_trait]
pub trait AIGatewayService: Send + Sync {
    async fn submit_request(&self, req: AIRequest) -> Result<AIResponse, AIGatewayError>;
    async fn stream_response(&self, req: AIRequest) -> Result<Receiver<AIChunk>, AIGatewayError>;
    async fn cancel_request(&self, id: RequestId) -> Result<(), AIGatewayError>;
    async fn get_session_status(&self, id: SessionId) -> Result<SessionStatus, AIGatewayError>;
}
```

---

## 7. Agent Orchestration Bounded Context

**Module:** `nexus-agent`
**Schema:** `agent_` prefix
**Purpose:** AI agent lifecycle, planning, tool selection, execution

### Aggregate: AgentExecution

```
AgentExecution (Aggregate Root)
├── Task
│   ├── TaskId
│   ├── Description
│   ├── Priority
│   └── Status
├── Plan
│   ├── Steps[]
│   └── Dependencies
├── ToolExecution[]
│   ├── ToolName
│   ├── Parameters
│   ├── Result
│   └── Status
└── Result
    ├── Output
    ├── TokensUsed
    └── LatencyMs
```

### Agent Routing Logic (In-Process)

```
User Request
    |
    v
nexus-ai-gateway → submit_request()
    |
    v
nexus-agent → start_execution()
    |
    v
Planner Agent (Ollama: lfm2.5-thinking:1.2b)
    |
    v
Intent Classification
    |
    ├── Coding     → Qwen2.5-Coder:3B
    ├── Security   → WhiteRabbitNeo:7B
    ├── Image      → Qwen3-VL:4B  (→ nexus-vision trait)
    ├── Document   → Command-R:7B (→ nexus-rag trait)
    ├── Business   → Llama3.1:7B
    └── General    → Phi-4-Mini:3.8B
```

---

## 8. RAG Knowledge Bounded Context

**Module:** `nexus-rag`
**Schema:** `rag_` prefix
**Purpose:** Enterprise knowledge intelligence

### Aggregate: KnowledgeDocument

```
KnowledgeDocument (Aggregate Root)
├── Metadata
│   ├── DocumentId
│   ├── TenantId
│   ├── FileName
│   ├── FileType
│   ├── Status
│   └── Tags[]
├── Chunks[]
│   ├── ChunkId
│   ├── Content
│   ├── Position
│   └── Embedding (vector(768))
└── AccessControl
    ├── DocumentSetId
    └── Classification
```

---

## 9. Vision Intelligence Bounded Context

**Module:** `nexus-vision`
**Schema:** `vision_` prefix
**Model:** `qwen3-vl:4b` (Ollama)

### Aggregate: VisionAnalysis

```
VisionAnalysis (Aggregate Root)
├── Image
│   ├── ImageId
│   ├── StoragePath
│   └── FileType
├── Detection
│   ├── Description
│   ├── Confidence
│   └── DetectedObjects[]
└── Extraction
    ├── Text (OCR)
    ├── StructuredData
    └── Metadata
```

---

## 10. SQL Intelligence Bounded Context

**Module:** `nexus-sql-agent`
**Schema:** `sql_` prefix
**Purpose:** Natural language business intelligence

### Aggregate: QueryExecution

```
QueryExecution (Aggregate Root)
├── GeneratedSQL
│   ├── RawSQL
│   ├── ParsedAST
│   └── ValidationStatus
├── ValidationResult
│   ├── IsSafe (read-only check)
│   ├── BlockedOperations
│   └── Permissions
└── ResultSet
    ├── Columns[]
    ├── Rows[]
    └── Summary
```

### SQL Safety Rules

**Allowed:** SELECT, JOIN, GROUP BY, ORDER BY, COUNT, SUM, AVG
**Blocked:** DELETE, UPDATE, DROP, ALTER, TRUNCATE, INSERT

---

## 11. Memory Bounded Context

**Module:** `nexus-memory`
**Schema:** `memory_` prefix + Redis
**Purpose:** Maintain AI memory across sessions

### Aggregate: MemoryProfile

```
MemoryProfile (Aggregate Root)
├── ShortTermMemory (Redis, TTL: 24h)
│   ├── CurrentConversation
│   ├── ActiveTasks
│   └── TemporaryContext
├── LongTermMemory (PostgreSQL + pgvector, permanent)
│   ├── UserPreferences
│   ├── PastConversations
│   └── ImportantFacts
└── OrganizationalMemory (Apache AGE)
    ├── EntityRelationships
    └── BusinessKnowledge
```

---

## 12. Workflow Bounded Context

**Module:** `nexus-workflow`
**Schema:** `workflow_` prefix
**Purpose:** Business automation and approvals

### Aggregate: WorkflowInstance

```
WorkflowInstance (Aggregate Root)
├── WorkflowDefinition
│   ├── Name
│   ├── Steps[]
│   └── Triggers
├── Steps[]
│   ├── StepId
│   ├── Type (ai_task, approval, notification, api_call, condition)
│   ├── Status
│   └── Assignee
├── Approvals[]
└── Actions[]
```

---

## 13. Security Intelligence Bounded Context

**Module:** `nexus-security-ai`
**Schema:** `security_` prefix
**Model:** `whiterabbitneo:7b` (Ollama)

### Aggregate: SecurityAnalysis

```
SecurityAnalysis (Aggregate Root)
├── Finding[]
│   ├── Severity (CRITICAL, HIGH, MEDIUM, LOW, INFO)
│   ├── Category
│   ├── Description
│   └── Recommendation
├── Recommendation[]
│   ├── Priority
│   ├── Action
│   └── Impact
└── RiskScore
    ├── Overall
    └── Breakdown[]
```

---

## 14. Audit Bounded Context

**Module:** `nexus-audit`
**Schema:** `audit_` prefix
**Purpose:** Complete AI activity tracking for compliance

### Domain Events (All Via NATS)

| Event | Description |
|---|---|
| `aeroxe.audit.ai.request` | AI interaction recorded |
| `aeroxe.audit.data.access` | Data access tracked |
| `aeroxe.audit.tool.execution` | Tool call recorded |
| `aeroxe.audit.security.event` | Security event tracked |

All other modules publish to these subjects. `nexus-audit` subscribes and persists to partitioned tables.

---

## 15. Domain Event Architecture

### Synchronous Events (In-Process)

For events that must be handled immediately within a transaction:

```rust
// Module emits event within its aggregate method
pub struct AgentExecuted {
    pub execution_id: ExecutionId,
    pub result: AgentResult,
    pub timestamp: ChronoDateTime,
}

// Event handler is called synchronously via a trait
pub trait AgentEventHandler: Send + Sync {
    fn handle(&self, event: &AgentExecuted);
}
```

### Asynchronous Events (NATS JetStream)

For cross-module notifications and background processing:

| Subject | Publisher | Consumers |
|---|---|---|
| `aeroxe.rag.document.uploaded` | nexus-rag | nexus-rag (self) |
| `aeroxe.rag.document.processed` | nexus-rag | nexus-gateway (status updates) |
| `aeroxe.agent.completed` | nexus-agent | nexus-ai-gateway, nexus-audit |
| `aeroxe.workflow.step.completed` | nexus-workflow | nexus-notification, nexus-audit |
| `aeroxe.security.threat.detected` | nexus-security-ai | nexus-audit, nexus-notification |

---

## 16. Final DDD Module Map

```
AeroXe Nexus AI — Modular Monolith

+---------------------------------------------------------------+
|                     Binary: aeroxe-nexus                        |
|                                                               |
|  Core Domain Modules:                                          |
|  ├── nexus-ai-gateway    (AI request lifecycle)                |
|  ├── nexus-agent         (Agent orchestration)                 |
|  ├── nexus-rag           (Document knowledge)                  |
|  ├── nexus-vision        (Image intelligence)                  |
|  ├── nexus-sql-agent     (Natural language SQL)                |
|  └── nexus-security-ai   (Security intelligence)               |
|                                                               |
|  Supporting Domain Modules:                                    |
|  ├── nexus-identity      (IAM, RBAC/ABAC)                     |
|  ├── nexus-memory        (AI memory system)                    |
|  ├── nexus-workflow      (Business automation)                 |
|  └── nexus-audit         (Compliance logging)                  |
|                                                               |
|  Infrastructure Modules:                                       |
|  ├── nexus-gateway       (API Gateway — HTTP/WS + middleware) |
|  ├── nexus-model-registry (Model management)                  |
|  ├── nexus-notification  (Email, WhatsApp, push)               |
|  ├── nexus-config        (Dynamic configuration)               |
|  └── nexus-ecosystem     (AeroXe product connectors)          |
+---------------------------------------------------------------+
```
