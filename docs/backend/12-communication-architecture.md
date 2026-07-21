# AeroXe Nexus AI — Communication Architecture

## Trait-Based Internal Calls + NATS JetStream Event-Driven Design

---

## 1. Communication Strategy Overview

AeroXe Nexus AI follows a **hybrid communication architecture** optimized for modular monoliths:

```
                     External World
                           |
                  REST / WebSocket / HTTPS
                           |
               +---------------------------+
               |    nexus-gateway Module    |
               |    (axum HTTP/WS Server)   |
               +---------------------------+
                           |
          ======================================
            Internal Communication (in-process)
          ======================================
              Synchronous          Asynchronous
              Trait Methods          NATS JetStream
                  |                       |
          Module-to-Module         Event-Driven Flow
          (direct fn calls)        (background jobs)
          ======================================
```

---

## 2. Communication Rules

| Layer | Protocol | Usage |
|---|---|---|
| External (Web/Mobile) | HTTPS REST | Client API |
| Real-time Chat | WebSocket | Token streaming |
| **Internal Synchronous** | **Rust trait methods** | Module-to-module (in-process) |
| Internal Asynchronous | NATS JetStream | Events, background jobs |
| AI Runtime | Ollama HTTP API | Model inference |
| External gRPC (optional) | tonic | SDK / partner integrations |

---

## 3. Trait-Based Internal Communication (Replaces gRPC)

### Design Principles

Every module communicates with other modules through **Rust trait interfaces**:

1. **Modules depend on traits, not implementations** — enables testing with mocks
2. **Traits are defined in each module** — in `src/interfaces/api.rs`
3. **Requests are in-process** — `Arc<dyn Trait>` references wired at startup
4. **Strong typing** — all inputs/outputs are typed Rust structs
5. **No serialization overhead** — direct struct passing
6. **Compile-time safety** — missing methods are compile errors

### Trait Definition Pattern

```rust
// nexus-rag/src/interfaces/api.rs
#[async_trait]
pub trait RagService: Send + Sync {
    async fn search(&self, query: SearchQuery) -> Result<SearchResults, RagError>;
    async fn upload_document(&self, request: UploadRequest) -> Result<DocumentStatus>;
}

// nexus-agent/src/agent_service.rs — uses RagService as dependency
pub struct AgentServiceImpl {
    rag: Arc<dyn RagService>,
    memory: Arc<dyn MemoryService>,
    ollama: OllamaClient,
}

#[async_trait]
impl AgentService for AgentServiceImpl {
    async fn start_execution(&self, request: StartAgentRequest) -> Result<ExecutionResponse> {
        // Call RAG synchronously — no gRPC, no serialization
        let docs = self.rag.search(SearchQuery {
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

### Wiring at Application Startup

```rust
// src/main.rs
async fn main() {
    // Shared infrastructure
    let db = PgPool::connect(&config.database_url).await?;
    let redis = RedisClient::open(config.redis_url)?;
    let nats = NatsClient::connect(config.nats_url).await?;
    let ollama = OllamaClient::new(config.ollama_url);

    // Instantiate module implementations (all in same process)
    let identity: Arc<dyn IdentityService> = Arc::new(
        nexus_identity::IdentityServiceImpl::new(db.clone(), redis.clone())
    );
    let rag: Arc<dyn RagService> = Arc::new(
        nexus_rag::RagServiceImpl::new(db.clone(), nats.clone(), ollama.clone())
    );
    let memory: Arc<dyn MemoryService> = Arc::new(
        nexus_memory::MemoryServiceImpl::new(db.clone(), redis.clone(), ollama.clone())
    );
    let agent: Arc<dyn AgentService> = Arc::new(
        nexus_agent::AgentServiceImpl::new(rag.clone(), memory.clone(), ollama.clone())
    );

    // Build gateway
    let app = nexus_gateway::build_router(AppState {
        identity, agent, rag, memory, // ...
    });

    axum::serve(listener, app).await?;
}
```

### Performance Comparison

| Aspect | Microservice (gRPC) | Modular Monolith (Trait) |
|---|---|---|
| Latency per call | 2-5ms (network + serialization) | < 1μs (trait vtable dispatch) |
| Overhead | Protobuf encode/decode | Zero — direct struct passing |
| Error handling | gRPC status codes | Rust `Result<T, E>` |
| Testing | Need running services | `Mockall` mocks |
| Type safety | Protobuf codegen | Rust compiler |

---

## 4. When to Use NATS vs Trait Calls

| Scenario | Use | Reason |
|---|---|---|
| User requests AI chat | Trait method | Need immediate response |
| Agent needs RAG data | Trait method | Synchronous data needed for reasoning |
| Agent needs memory | Trait method | Synchronous context retrieval |
| Agent execution completes | NATS | Other modules need to react asynchronously |
| Document uploaded | NATS | Long-running processing, don't block client |
| Audit event | NATS | Fire-and-forget, must not impact latency |
| Notification send | NATS | Background delivery, retry on failure |
| Workflow step completed | NATS | Decoupled step orchestration |
| Config change broadcast | NATS | All modules must eventually reconfigure |

---

## 5. Module Service Dependency Graph

```
                    nexus-gateway
                         |
        +----------------+----------------+----------------+
        |                |                |                |
   nexus-identity  nexus-ai-gateway  nexus-model-registry  |
        |                |                                   |
        |          nexus-agent                                |
        |                |       |        |        |          |
        |                v       v        v        v          |
        |          nexus-rag  nexus-  nexus-  nexus-memory   |
        |                     vision  sql-agent               |
        |                                                     |
        +-------------------+-----+------+------------------+
                            |     |      |
                     nexus-workflow  nexus-security-ai
                            |              |
                            v              v
                     nexus-notification  nexus-audit
```

---

## 6. Module API Trait Catalogue

### nexus-identity

```rust
#[async_trait]
pub trait IdentityService: Send + Sync {
    async fn authenticate(&self, req: AuthRequest) -> Result<AuthResponse>;
    async fn verify_token(&self, token: &str) -> Result<JWTClaims>;
    async fn check_permission(&self, req: PermissionRequest) -> Result<bool>;
    async fn validate_tenant(&self, tenant_id: TenantId) -> Result<Tenant>;
    async fn create_user(&self, req: CreateUserRequest) -> Result<User>;
}
```

### nexus-ai-gateway

```rust
#[async_trait]
pub trait AIGatewayService: Send + Sync {
    async fn submit_request(&self, req: AIRequest) -> Result<AIResponse>;
    async fn stream_response(&self, req: AIRequest) -> Result<Receiver<AIChunk>>;
    async fn cancel_request(&self, id: RequestId) -> Result<()>;
    async fn get_session_status(&self, id: SessionId) -> Result<SessionStatus>;
}
```

### nexus-agent

```rust
#[async_trait]
pub trait AgentService: Send + Sync {
    async fn start_execution(&self, req: StartAgentRequest) -> Result<ExecutionResponse>;
    async fn get_execution_status(&self, id: ExecutionId) -> Result<ExecutionStatus>;
    async fn stream_execution(&self, req: StreamRequest) -> Result<Receiver<ExecutionEvent>>;
    async fn cancel_execution(&self, id: ExecutionId) -> Result<()>;
}
```

### nexus-rag

```rust
#[async_trait]
pub trait RagService: Send + Sync {
    async fn search(&self, query: SearchQuery) -> Result<SearchResults>;
    async fn upload_document(&self, req: UploadRequest) -> Result<DocumentStatus>;
    async fn get_document_status(&self, id: DocumentId) -> Result<Option<DocumentStatus>>;
    async fn delete_document(&self, id: DocumentId) -> Result<()>;
}
```

### nexus-vision

```rust
#[async_trait]
pub trait VisionService: Send + Sync {
    async fn analyze_image(&self, req: ImageRequest) -> Result<ImageAnalysisResponse>;
    async fn extract_text(&self, req: ImageRequest) -> Result<OCRResponse>;
    async fn troubleshoot_device(&self, req: DeviceImageRequest) -> Result<TroubleshootResponse>;
}
```

### nexus-sql-agent

```rust
#[async_trait]
pub trait SQLAgentService: Send + Sync {
    async fn generate_query(&self, req: QueryRequest) -> Result<SQLResponse>;
    async fn execute_query(&self, req: SQLRequest) -> Result<ResultResponse>;
    async fn test_connection(&self, req: TestConnectionRequest) -> Result<ConnectionStatus>;
}
```

### nexus-memory

```rust
#[async_trait]
pub trait MemoryService: Send + Sync {
    async fn store(&self, req: StoreMemoryRequest) -> Result<()>;
    async fn search(&self, req: SearchMemoryRequest) -> Result<Vec<MemoryItem>>;
    async fn get_conversation_context(&self, session_id: SessionId) -> Result<Vec<Message>>;
    async fn clear_session(&self, session_id: SessionId) -> Result<()>;
}
```

### nexus-workflow

```rust
#[async_trait]
pub trait WorkflowService: Send + Sync {
    async fn start_workflow(&self, req: StartWorkflowRequest) -> Result<WorkflowResponse>;
    async fn get_status(&self, id: WorkflowId) -> Result<WorkflowStatus>;
    async fn approve_step(&self, req: ApproveRequest) -> Result<()>;
    async fn cancel_workflow(&self, id: WorkflowId) -> Result<()>;
}
```

### nexus-security-ai

```rust
#[async_trait]
pub trait SecurityService: Send + Sync {
    async fn analyze_code(&self, req: CodeReviewRequest) -> Result<CodeReviewResponse>;
    async fn scan_security(&self, req: SecurityScanRequest) -> Result<SecurityReport>;
    async fn scan_prompt(&self, prompt: &str) -> Result<PromptScanResult>;
}
```

### nexus-audit

```rust
#[async_trait]
pub trait AuditService: Send + Sync {
    async fn log_event(&self, event: AuditEvent) -> Result<()>;
    async fn query_events(&self, query: AuditQuery) -> Result<Vec<AuditEvent>>;
    async fn generate_report(&self, req: ReportRequest) -> Result<Report>;
}
```

---

## 7. NATS JetStream Architecture

### Subject Naming Standard

Format: `aeroxe.<module>.<event>`

### All Subjects

| Subject | Module | Description |
|---|---|---|
| `aeroxe.ai.request.created` | AI Gateway | New AI request |
| `aeroxe.ai.response.generated` | AI Gateway | Response ready |
| `aeroxe.ai.failed` | AI Gateway | Request failed |
| `aeroxe.agent.started` | Agent | Agent execution started |
| `aeroxe.agent.completed` | Agent | Agent execution done |
| `aeroxe.agent.failed` | Agent | Agent execution failed |
| `aeroxe.agent.tool.executed` | Agent | Tool call made |
| `aeroxe.rag.document.uploaded` | RAG | Document received |
| `aeroxe.rag.document.processed` | RAG | Processing done |
| `aeroxe.rag.embedding.created` | RAG | Embeddings stored |
| `aeroxe.rag.knowledge.updated` | RAG | Knowledge base modified |
| `aeroxe.vision.image.received` | Vision | Image uploaded |
| `aeroxe.vision.analysis.completed` | Vision | Analysis done |
| `aeroxe.workflow.started` | Workflow | Workflow started |
| `aeroxe.workflow.step.completed` | Workflow | Step finished |
| `aeroxe.workflow.completed` | Workflow | All steps complete |
| `aeroxe.workflow.failed` | Workflow | Workflow error |
| `aeroxe.security.scan.started` | Security | Scan initiated |
| `aeroxe.security.threat.detected` | Security | Threat found |
| `aeroxe.audit.*` | Audit | All audit events |
| `aeroxe.identity.*` | Identity | User/tenant events |
| `aeroxe.memory.*` | Memory | Memory lifecycle events |
| `aeroxe.gateway.*` | Gateway | Gateway operational events |
| `aeroxe.config.*` | Config | Configuration changes |

---

## 8. Event Schema Standard

```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "AgentCompleted",
  "timestamp": "2026-07-21T12:00:00Z",
  "tenant_id": 1,
  "module": "nexus-agent",
  "version": "1.0",
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

## 9. JetStream Stream Design

| Stream Name | Subjects | Retention | Replication |
|---|---|---|---|
| `AI_EVENTS` | `aeroxe.ai.*` | 7 days | 3 |
| `AGENT_EVENTS` | `aeroxe.agent.*` | 30 days | 3 |
| `RAG_EVENTS` | `aeroxe.rag.*` | 14 days | 1 |
| `AUDIT_EVENTS` | `aeroxe.audit.*` | 365 days | 3 |
| `WORKFLOW_EVENTS` | `aeroxe.workflow.*` | 30 days | 1 |
| `SECURITY_EVENTS` | `aeroxe.security.*` | 365 days | 3 |

---

## 10. Request Flow Example

**User asks:** "Why is customer internet slow?"

```
User → HTTP POST /api/v1/ai/chat
  → nexus-gateway (auth, rate-limit)
    → nexus-ai-gateway::submit_request() [trait call]
      → nexus-agent::start_execution() [trait call]
        → Ollama: LFM2.5 Thinking (intent detection)
          → Plan: check customer, check network, search knowledge
        → nexus-rag::search() [trait call] — document knowledge
        → nexus-memory::search() [trait call] — past conversations
        → Ollama: Command-R 7B (generate final answer)
      ← Response + audit event (NATS: aeroxe.audit.ai.request)
    ← HTTP 200 + JSON response
```

Note: **Every arrow is an in-process Rust trait method call**, not a network hop. The entire flow completes in < 3 seconds.

---

## 11. Streaming Response Architecture

```
Client WebSocket
    |
    v
nexus-gateway (axum WebSocket handler)
    |
    v
nexus-ai-gateway::stream_response()
    |  → returns tokio::sync::mpsc::Receiver<AIChunk>
    v
nexus-agent::stream_execution()
    |  → returns tokio::sync::mpsc::Receiver<ExecutionEvent>
    v
Ollama HTTP streaming API
    |
    v
Token stream → Receiver → WebSocket → Client
```

All streams use **tokio channels** for zero-copy token relay between modules.

---

## 12. Security Requirements

### Internal Trait Calls

| Requirement | Implementation |
|---|---|
| Authentication | JWT validated by nexus-gateway, claims attached to request |
| Authorization | nexus-identity::check_permission() called by gateway |
| Tenant Isolation | tenant_id propagated via RequestContext struct |
| Rate Limiting | Token bucket in gateway middleware |

### NATS Security

| Requirement | Implementation |
|---|---|
| TLS | All NATS connections encrypted |
| Account Isolation | Separate accounts per module |
| Subject Permissions | Publish/subscribe ACLs |
| Authentication | NKey or JWT-based auth |

---

## 13. Testing Communication Contracts

### Module Boundary Tests (TDD)

```rust
/// Contract: nexus-agent must be able to call nexus-rag::search()
/// This test validates the trait interface, not a specific implementation.
#[tokio::test]
async fn test_agent_rag_integration() {
    let rag = MockRagService::new();
    // The mock implements the same trait as the real service
    let agent = AgentServiceImpl::new(Arc::new(rag), /* ... */);

    let result = agent.start_execution(/* ... */).await;
    assert!(result.is_ok());
}
```

### NATS Contract Tests

```rust
#[tokio::test]
async fn test_agent_completed_event_is_published() {
    let nats = nats_test_server().await;
    let agent = AgentServiceImpl::new(/* ... with nats */);

    agent.start_execution(/* ... */).await;

    // Verify event was published
    let msg = nats.subscribe("aeroxe.agent.completed").await?;
    // ...
}
```
