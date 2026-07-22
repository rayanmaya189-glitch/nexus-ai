# AeroXe Nexus AI — DDD Domain Design

## Domain-Driven Design + Modular Monolith + Aggregate Design

---

## 1. DDD Architecture Overview

AeroXe Nexus AI is designed using **Domain-Driven Design** principles organized as a **modular monolith**. The system is divided into independent **Bounded Contexts**, each owning its business logic, database schema (via SeaORM entities), and public API. Modules communicate via **gRPC** (synchronous request-response) and **NATS events** (async fire-and-forget) with Protobuf payloads.

### Key Difference from Microservices

| Aspect | Microservice Architecture | Modular Monolith (This Project) |
|---|---|---|
| Communication | gRPC over network | gRPC (in-process) + NATS (async) |
| Database | Separate DB per service | Shared PostgreSQL, schema per module via SeaORM |
| ORM | N/A (each service chooses) | SeaORM (unified across all modules) |
| Deployment | N containers | 1 binary |
| Testing | Service-level integration tests | Module-level + full binary tests |
| Latency | 2-5ms per gRPC call | < 2ms per in-process gRPC call |
| Extractability | N/A | Any module can be extracted to a standalone service later if needed |

---

## 2. Core Domain Classification

| Domain | Type | Module Name | Schema Prefix |
|---|---|---|---|
| Agent Orchestration | **Core Domain** | `agent` | `agent_` |
| AI Gateway | **Core Domain** | `ai-gateway` | `ai_` |
| RAG Intelligence | **Core Domain** | `rag` | `rag_` |
| Vision Intelligence | **Core Domain** | `vision` | `vision_` |
| SQL Intelligence | **Core Domain** | `sql-agent` | `sql_` |
| Security Intelligence | **Core Domain** | `security` | `security_` |
| Identity | Supporting | `identity` | `identity_` |
| Customer | Supporting | `customer` | `customer_` |
| Memory | Supporting | `memory` | `memory_` |
| Audit | Supporting | `audit` | `audit_` |
| Workflow | Supporting | `workflow` | `workflow_` |

### Voice/AI Channel Modules (NEW)

| Domain | Type | Module Name | Schema Prefix |
|---|---|---|---|
| Telephony | **Core Domain** | `telephony` | `telephony_` |
| Conversation Management | **Core Domain** | `conversation` | `conversation_` |
| Speech-to-Text | **Core Domain** | `stt` | `stt_` |
| Text-to-Speech | **Core Domain** | `tts` | `tts_` |
| Analytics & Intelligence | Supporting | `analytics` | `analytics_` |
| Integration (Webhooks) | Supporting | `webhook` | `webhook_` |
| Outbound Communication | Supporting | `outbound` | `outbound_` |

### Infrastructure Modules

| Module | Purpose | Schema Prefix |
|---|---|---|
| `gateway` | API Gateway (axum HTTP/WS) | — (stateless) |
| `model-registry` | Ollama model management | `models_` |
| `notification` | Email, WhatsApp, push | `notification_` |
| `config` | Dynamic configuration | `config_` |
| `ecosystem` | AeroXe product connectors | `eco_` |

---

## 3. Module Internal Code Structure (DDD Layers)

Every module follows a strict layered architecture within `src/modules/<name>/`:

```
src/modules/<name>/                # Module root
├── mod.rs                        # Public API: trait definitions + re-exports
├── domain/                       # 🟢 DDD Layer: Business logic
│   ├── mod.rs
│   ├── aggregates/               # Aggregate roots with invariants
│   │   └── <aggregate>/
│   │       ├── <aggregate>.rs    # Aggregate root struct
│   │       ├── <entity>.rs       # Nested entities
│   │       └── tests/            # Domain tests
│   ├── entities/                 # Mutable domain objects (with IDs)
│   ├── value_objects/            # Immutable validated types (no IDs)
│   ├── events/                   # Domain event structs + serialization
│   └── rules/                    # Business rules / invariants
├── application/                  # 🟡 DDD Layer: Use cases / orchestration
│   ├── mod.rs
│   ├── commands/                 # Command structs (what you tell the system to do)
│   ├── queries/                  # Query structs (what you ask the system)
│   ├── handlers/                 # Command/Query handler implementations
│   └── services/                 # Application services (orchestration)
├── infrastructure/               # 🔵 DDD Layer: Technical concerns
│   ├── mod.rs
│   ├── repository/               # SeaORM repositories (entity models)
│   ├── ollama/                   # Ollama HTTP client adapter
│   ├── messaging/                # NATS publisher/subscriber adapter
│   │   ├── publishers/
│   │   └── subscribers/
│   └── security/                 # Security (JWT, hashing)
├── api/                          # 🟣 Interface Adapters
│   ├── mod.rs
│   ├── http/                     # HTTP controllers (axum handlers)
│   └── external/                 # External adapters (gRPC, partner SDKs)
├── migrations/                   # SeaORM migration files
└── tests/                        # Module tests
    ├── unit/                     # 🔴 Domain unit tests (no infra)
    ├── integration/              # 🟠 Integration tests (DB via SeaORM)
    ├── contract/                 # 🟢 Module boundary contract tests
    └── e2e/                      # 🔵 Full flow tests (all modules)
```

---

## 4. Module Public API Trait Pattern

Each module exposes its public API as Rust traits (internal implementation detail). Inter-service communication uses **gRPC** (synchronous) or **NATS** (async). The trait implementations serve as the service layer behind gRPC handlers.

```rust
// src/modules/rag/src/domain/interfaces/api.rs
#[async_trait]
pub trait RagService: Send + Sync {
    async fn search(&self, query: SearchQuery) -> Result<SearchResults, RagError>;
    async fn upload_document(&self, request: UploadRequest) -> Result<DocumentStatus>;
    async fn get_document_status(&self, id: DocumentId) -> Result<Option<DocumentStatus>>;
    async fn delete_document(&self, id: DocumentId) -> Result<()>;
}

// src/modules/agent/application/services/agent_service.rs — uses RagService via gRPC
pub struct AgentServiceImpl {
    rag_client: RagServiceClient,  // gRPC client
    memory_client: MemoryServiceClient,  // gRPC client
    ollama: OllamaClient,
}

#[async_trait]
impl AgentService for AgentServiceImpl {
    async fn start_execution(&self, request: StartAgentRequest) -> Result<ExecutionResponse> {
        // Call RAG via gRPC — Protobuf request/response
        let docs = self.rag_client.search(SearchQuery {
            query: request.task.clone(),
            tenant_id: request.tenant_id,
            limit: 5,
        }).await?;

        // Call Ollama directly
        let plan = self.ollama.chat(/* ... */).await?;

        Ok(ExecutionResponse { /* ... */ })
    }
}
```

The binary's `main.rs` wires all services together, starting gRPC servers for each module:

```rust
// src/main.rs (binary entry point)
#[tokio::main]
async fn main() {
    let db = init_db().await;          // SeaORM DatabaseConnection
    let redis = init_redis().await;
    let nats = init_nats().await;
    let ollama = OllamaClient::new(config.ollama_url);

    // Start gRPC servers for each service module
    let identity_server = identity::start_grpc_server(db.clone(), redis.clone());
    let customer_server = customer::start_grpc_server(db.clone(), nats.clone());
    let memory_server = memory::start_grpc_server(db.clone(), redis.clone(), ollama.clone());
    let rag_server = rag::start_grpc_server(db.clone(), nats.clone(), ollama.clone());
    let agent_server = agent::start_grpc_server(/* ... */);

    // Start NATS subscribers for async event processing
    let audit_subscriber = audit::start_nats_subscriber(nats.clone());

    // Start API Gateway (HTTP + WebSocket)
    let gateway = gateway::start_http_server(/* ... */);

    // Run all services concurrently
    tokio::join!(
        identity_server,
        customer_server,
        memory_server,
        rag_server,
        agent_server,
        audit_subscriber,
        gateway,
    );
}
```

---

## 5. SeaORM Repository Pattern

All database access uses **SeaORM entities and models** — no raw SQL.

```rust
// src/modules/agent/infrastructure/repository/agent_repo.rs
use sea_orm::*;
use sea_orm::entity::prelude::*;

// SeaORM entity (generated or handwritten)
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "agents", schema_name = "agent")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub name: String,
    pub agent_type: String,
    pub model: String,
    pub system_prompt: Option<String>,
    pub capabilities: Option<Json>,
    pub status: String,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(has_many = "super::executions::Entity")]
    Executions,
}

impl ActiveModelBehavior for ActiveModel {}

// Repository implementation
#[async_trait]
impl AgentRepository for PostgresAgentRepository {
    async fn save(&self, agent: Agent) -> Result<AgentId, RepositoryError> {
        let model = ActiveModel {
            name: Set(agent.name),
            agent_type: Set(agent.agent_type.to_string()),
            model: Set(agent.model),
            system_prompt: Set(agent.system_prompt),
            capabilities: Set(agent.capabilities.map(Json)),
            status: Set(agent.status.to_string()),
            ..Default::default()
        };
        let result = model.insert(&self.db).await?;
        Ok(AgentId(result.id))
    }

    async fn find_by_id(&self, id: AgentId) -> Result<Option<Agent>, RepositoryError> {
        let model = Entity::find_by_id(id.0)
            .one(&self.db)
            .await?;
        Ok(model.map(Agent::from))
    }

    async fn find_by_tenant(&self, tenant_id: TenantId) -> Result<Vec<Agent>, RepositoryError> {
        let models = Entity::find()
            .filter(ModelColumn::TenantId.eq(tenant_id.0))
            .all(&self.db)
            .await?;
        Ok(models.into_iter().map(Agent::from).collect())
    }
}
```

---

## 6. Identity Bounded Context

**Module:** `identity`
**Location:** `src/modules/identity/`
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
| Session | SessionId, UserId, Token, ExpiresAt |

### Value Objects

| Value Object | Rust Type | Validation |
|---|---|---|
| `UserId` | `i64` | Positive |
| `TenantId` | `i64` | Positive |
| `EmailAddress` | `String` | Regex validated (`validator` crate) |
| `Permission` | `String` | Format: `{resource}.{action}` |
| `PasswordHash` | `String` | bcrypt verified |
| `JWTToken` | `String` | RS256 signed + exp check |

### Domain Events (Versioned NATS Subjects)

| Event | Trigger | NATS Subject |
|---|---|---|
| New user registration | User created | `aeroxe.v1.identity.user.created` |
| Profile change | User updated | `aeroxe.v1.identity.user.updated` |
| Role assignment | Role assigned | `aeroxe.v1.identity.role.assigned` |
| Permission modification | Permission changed | `aeroxe.v1.identity.permission.changed` |

### Public API Trait

```rust
// src/modules/identity/api/mod.rs
#[async_trait]
pub trait IdentityService: Send + Sync {
    async fn authenticate(&self, req: AuthRequest) -> Result<AuthResponse, IdentityError>;
    async fn verify_token(&self, token: &str) -> Result<JWTClaims, IdentityError>;
    async fn check_permission(&self, req: PermissionRequest) -> Result<bool, IdentityError>;
    async fn validate_tenant(&self, tenant_id: TenantId) -> Result<Tenant, IdentityError>;
    async fn create_user(&self, req: CreateUserRequest) -> Result<User, IdentityError>;
    async fn get_user(&self, id: UserId) -> Result<Option<User>, IdentityError>;
    async fn assign_role(&self, req: AssignRoleRequest) -> Result<(), IdentityError>;
}
```

---

## 7. Customer Bounded Context

**Module:** `customer`
**Location:** `src/modules/customer/`
**Schema:** `customer_` prefix in shared PostgreSQL
**Purpose:** Manage customers, profiles, KYC data, addresses, customer lifecycle

### Aggregate: Customer

```
Customer (Aggregate Root)
├── Profile
│   ├── CustomerId
│   ├── Name (value object)
│   ├── EmailAddress (value object)
│   ├── PhoneNumber (value object)
│   └── Addresses[] (entities)
├── Status (value object / enum)
│   ├── Active
│   ├── Suspended
│   ├── Inactive
│   └── Archived
├── KYC
│   ├── KYC Status
│   ├── Document References
│   └── Verification Date
└── Metadata
    ├── Tags[]
    ├── CustomFields (JSON)
    └── Notes[]
```

### Domain Events (Versioned NATS Subjects)

| Event | Trigger | NATS Subject |
|---|---|---|
| Customer created | New customer registered | `aeroxe.v1.customer.customer.created` |
| Customer activated | Status changed to active | `aeroxe.v1.customer.customer.activated` |
| Customer suspended | Status changed to suspended | `aeroxe.v1.customer.customer.suspended` |
| Customer profile updated | Profile changed | `aeroxe.v1.customer.customer.updated` |

### Public API Trait

```rust
// src/modules/customer/api/mod.rs
#[async_trait]
pub trait CustomerService: Send + Sync {
    async fn create_customer(&self, req: CreateCustomerRequest) -> Result<Customer, CustomerError>;
    async fn get_customer(&self, id: CustomerId, tenant_id: TenantId) -> Result<Option<Customer>, CustomerError>;
    async fn suspend_customer(&self, id: CustomerId, tenant_id: TenantId) -> Result<Customer, CustomerError>;
    async fn activate_customer(&self, id: CustomerId, tenant_id: TenantId) -> Result<Customer, CustomerError>;
    async fn update_customer(&self, req: UpdateCustomerRequest) -> Result<Customer, CustomerError>;
    async fn search_customers(&self, query: CustomerSearchQuery) -> Result<Vec<Customer>, CustomerError>;
}
```

---

## 8. AI Gateway Bounded Context

**Module:** `ai-gateway`
**Location:** `src/modules/ai-gateway/`
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
    ├── TargetAgent (via agent module)
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

## 9. Agent Orchestration Bounded Context

**Module:** `agent`
**Location:** `src/modules/agent/`
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
ai-gateway → submit_request()
    |
    v
agent → start_execution()
    |
    v
Planner Agent (Ollama: lfm2.5-thinking:1.2b)
    |
    v
Intent Classification
    |
    ├── Coding     → Qwen2.5-Coder:3B
    ├── Security   → WhiteRabbitNeo:7B
    ├── Image      → Qwen3-VL:4B  (→ vision trait)
    ├── Document   → Command-R:7B (→ rag trait)
    ├── Business   → Llama3.1:7B
    └── General    → Phi-4-Mini:3.8B
```

---

## 10. RAG Knowledge Bounded Context

**Module:** `rag`
**Location:** `src/modules/rag/`
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

## 11. Domain Event Architecture (Versioned)

### Synchronous Events (gRPC)

For events that must be handled immediately within a transaction, services use **gRPC** for synchronous request-response:

```rust
// Agent service calls RAG service via gRPC for document search
let results = rag_client.search(SearchRequest {
    query: task_description,
    tenant_id: tenant_id,
    limit: 5,
}).await?;
```

### Asynchronous Events (NATS JetStream — Protobuf Payloads)

For cross-module notifications and background processing — all NATS event payloads are **Protobuf messages** and subjects include `v1` version prefix:

| Subject | Publisher | Consumers |
|---|---|---|
| `aeroxe.v1.rag.document.uploaded` | rag | rag (self) |
| `aeroxe.v1.rag.document.processed` | rag | gateway (status updates) |
| `aeroxe.v1.agent.completed` | agent | ai-gateway, audit |
| `aeroxe.v1.workflow.step.completed` | workflow | notification, audit |
| `aeroxe.v1.security.threat.detected` | security | audit, notification |
| `aeroxe.v1.customer.customer.created` | customer | notification, audit |
| `aeroxe.v1.customer.customer.suspended` | customer | notification, workflow |
| `aeroxe.v1.identity.user.created` | identity | audit, notification |
| `aeroxe.v1.identity.role.assigned` | identity | audit |

### Event Envelope (Protobuf)

Every NATS event follows a standard Protobuf message serialized as JSON or binary:

```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "AgentCompleted",
  "version": "1.0",
  "api_version": "v1",
  "timestamp": "2026-07-21T12:00:00Z",
  "tenant_id": 1,
  "module": "agent",
  "data": {
    "execution_id": 12345,
    "agent": "customer-agent",
    "status": "success",
    "tokens_used": 1250,
    "latency_ms": 3500
  }
}
```

---

## 12. Final DDD Module Map

```
AeroXe Nexus AI — Modular Monolith
src/modules/

+---------------------------------------------------------------+
|                     Binary: aeroxe-nexus                        |
|                                                               |
|  Core Domain Modules:                                          |
|  ├── ai-gateway    (AI request lifecycle)               ai_   |
|  ├── agent         (Agent orchestration)                agent_|
|  ├── rag           (Document knowledge)                  rag_  |
|  ├── vision        (Image intelligence)                  vision|
|  ├── sql-agent     (Natural language SQL)                sql_  |
|  ├── security      (Security intelligence)               security_ |
|  ├── telephony     (Voice channel, SIP/WebRTC)          telephony_ |
|  ├── conversation  (State machine, context)             conversation_ |
|  ├── stt           (Speech-to-Text)                     stt_  |
|  └── tts           (Text-to-Speech)                     tts_  |
|                                                               |
|  Supporting Domain Modules:                                    |
|  ├── identity      (IAM, RBAC, ABAC, KYC)               identity_ |
|  ├── customer      (Customer management)                 customer_ |
|  ├── memory        (AI memory system)                    memory_ |
|  ├── workflow      (Business automation)                 workflow_ |
|  ├── audit         (Compliance logging)                  audit |
|  ├── analytics     (Metrics, reports, BI)                anly_ |
|  ├── webhook       (Event delivery, retry)               whk_  |
|  └── outbound      (Proactive calls, campaigns)         out_  |
|                                                               |
|  Infrastructure Modules:                                       |
|  ├── gateway       (API Gateway — HTTP/WS + middleware)   —    |
|  ├── model-registry (Model management)                   mode  |
|  ├── notification  (Email, WhatsApp, push)               notification_ |
|  ├── config        (Dynamic configuration)                conf |
|  └── ecosystem     (AeroXe product connectors)            eco_ |
+---------------------------------------------------------------+
```
