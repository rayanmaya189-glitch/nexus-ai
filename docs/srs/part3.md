# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 3 — Communication Architecture Design

## Rust Trait Interfaces (In-Process) + Versioned NATS JetStream Events + External gRPC (Optional)

---

# 1. Communication Strategy Overview

AeroXe Nexus AI follows a **hybrid communication architecture** optimized for modular monoliths.

The communication model:

```text
                     External World


                           |

                  REST / WebSocket / HTTPS (versioned: /api/v1/)


                           |

                    gateway Module


                           |

================================================

                 Internal Communication


================================================


        Synchronous              Asynchronous


    Rust Trait Interfaces     NATS JetStream (versioned)


     (in-process,             (aeroxe.v1.module.event)
      < 1μs dispatch)


             |                         |


     Module-to-Module        Event Driven Flow


================================================


            External gRPC (optional, versioned service names)

            tonic — for SDK / partner integrations only


```

---

# 2. Communication Rules

## External Communication

Used by:

* Web applications
* Mobile applications
* AeroXe products
* Third-party integrations

Protocol:

```
HTTPS REST (/api/v1/*)
WebSocket (/ws/v1/*)
```

---

## Internal Synchronous Communication

**Modular Monolith:** Modules communicate through **Rust trait interfaces** — no gRPC, no network.

```rust
// Example: agent module calls rag module
let docs = self.rag_service.search(SearchQuery {
    query: request.task,
    tenant_id: request.tenant_id,
    limit: 5,
}).await?;
```

Benefits:

| Aspect | gRPC (Microservice) | Trait Interface (Modular Monolith) |
|---|---|---|
| Latency | 2-5ms | < 1μs (vtable dispatch) |
| Serialization | Protobuf encode/decode | Zero — direct struct passing |
| Type safety | Protobuf codegen | Rust compiler |
| Testing | Need running services | Mockall mocks |

---

## Event Communication

Used for:

* Background jobs
* Notifications
* AI workflow
* Data synchronization
* **Reliable delivery via Outbox Pattern**

Protocol:

```
NATS JetStream (versioned subjects: aeroxe.v1.*)
```

---

# 3. Trait Interface Architecture

**Key Difference:** In the modular monolith, all modules are in the same binary. They communicate through Rust trait interfaces, not gRPC. This eliminates:

- Network latency
- Serialization overhead
- mTLS complexity (not needed in-process)
- Service discovery (not needed in-process)

---

# 3. Trait Interface Architecture

## Module Communication

Example:

```text
agent module


          |

          | Rust trait method call

          |

rag module


          |

          | Rust trait method call

          |

memory module

```

---

# 4. Trait Design Principles

Every module exposes its public API as Rust traits:

* Versioned trait methods (backward compatible)
* Strong typing
* Error standards (`Result<T, E>`)
* Authentication context in `RequestContext`

Example:

```rust
pub trait IdentityService: Send + Sync {
    async fn verify_token(&self, token: &str) -> Result<JWTClaims, IdentityError>;
    async fn check_permission(&self, req: PermissionRequest) -> Result<bool, IdentityError>;
}
```

---

# 5. Module API Trait Repository

All module traits live in each module's `api/mod.rs`:

```
src/modules/

├── identity/api/mod.rs      → IdentityService trait
├── customer/api/mod.rs      → CustomerService trait  ← NEW
├── agent/api/mod.rs         → AgentService trait
├── rag/api/mod.rs           → RagService trait
├── vision/api/mod.rs        → VisionService trait
├── memory/api/mod.rs        → MemoryService trait
├── sql-agent/api/mod.rs     → SQLAgentService trait
├── workflow/api/mod.rs      → WorkflowService trait
├── security/api/mod.rs      → SecurityService trait
├── audit/api/mod.rs         → AuditService trait

```

---

# 6. Common Request Context

File:

```
src/modules/common/request_context.rs
```

```rust
pub struct RequestContext {
    pub request_id: String,
    pub tenant_id: String,
    pub user_id: String,
    pub trace_id: String,
    pub api_version: String,    // e.g., "v1"
}

pub struct ErrorResponse {
    pub code: String,
    pub message: String,
    pub request_id: String,
    pub api_version: String,
    pub timestamp: String,
}
```

---

# 7. Identity Module API Trait

Module:

```
identity (src/modules/identity/)
```

Trait:

```rust
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

# 8. Customer Module API Trait (NEW)

Module:

```
customer (src/modules/customer/)
```

```rust
#[async_trait]
pub trait CustomerService: Send + Sync {
    async fn create_customer(&self, req: CreateCustomerRequest) -> Result<Customer, CustomerError>;
    async fn get_customer(&self, id: CustomerId, tenant_id: TenantId) -> Result<Option<Customer>, CustomerError>;
    async fn suspend_customer(&self, id: CustomerId, tenant_id: TenantId, reason: String) -> Result<Customer, CustomerError>;
    async fn activate_customer(&self, id: CustomerId, tenant_id: TenantId) -> Result<Customer, CustomerError>;
    async fn search_customers(&self, query: CustomerSearchQuery) -> Result<Vec<Customer>, CustomerError>;
}
```

---

# 9. AI Gateway Module API Trait

Module:

```
ai-gateway (src/modules/ai-gateway/)
```

```rust
#[async_trait]
pub trait AIGatewayService: Send + Sync {
    async fn submit_request(&self, req: AIRequest) -> Result<AIResponse, AIGatewayError>;
    async fn stream_response(&self, req: AIRequest) -> Result<Receiver<AIChunk>, AIGatewayError>;
    async fn cancel_request(&self, id: RequestId) -> Result<(), AIGatewayError>;
}
```

---

# 10. Agent Module API Trait

Module:

```
agent (src/modules/agent/)
```

```rust
#[async_trait]
pub trait AgentService: Send + Sync {
    async fn start_execution(&self, req: StartAgentRequest) -> Result<ExecutionResponse, AgentError>;
    async fn get_execution_status(&self, id: ExecutionId) -> Result<ExecutionStatus, AgentError>;
    async fn stream_execution(&self, req: StreamRequest) -> Result<Receiver<ExecutionEvent>, AgentError>;
}
```

---

# 11. RAG Module API Trait

Module:

```
rag (src/modules/rag/)
```

```rust
#[async_trait]
pub trait RagService: Send + Sync {
    async fn search(&self, query: SearchQuery) -> Result<SearchResults, RagError>;
    async fn upload_document(&self, req: UploadRequest) -> Result<DocumentStatus, RagError>;
    async fn get_document_status(&self, id: DocumentId) -> Result<Option<DocumentStatus>, RagError>;
}
```

---

# 12. Vision Module API Trait

Module:

```
vision (src/modules/vision/)
```

```rust
#[async_trait]
pub trait VisionService: Send + Sync {
    async fn analyze_image(&self, req: ImageRequest) -> Result<ImageAnalysisResponse, VisionError>;
    async fn extract_text(&self, req: ImageRequest) -> Result<OCRResponse, VisionError>;
}
```

---

# 13. SQL Agent Module API Trait

Module:

```
sql-agent (src/modules/sql-agent/)
```

```rust
#[async_trait]
pub trait SQLAgentService: Send + Sync {
    async fn generate_query(&self, req: QueryRequest) -> Result<SQLResponse, SQLError>;
    async fn execute_query(&self, req: SQLRequest) -> Result<ResultResponse, SQLError>;
}
```

---

# 14. Memory Module API Trait

Module:

```
memory (src/modules/memory/)
```

```rust
#[async_trait]
pub trait MemoryService: Send + Sync {
    async fn store(&self, req: StoreMemoryRequest) -> Result<(), MemoryError>;
    async fn search(&self, req: SearchMemoryRequest) -> Result<Vec<MemoryItem>, MemoryError>;
    async fn get_conversation_context(&self, session_id: SessionId) -> Result<Vec<Message>, MemoryError>;
}
```

---

# 15. Workflow Module API Trait

Module:

```
workflow (src/modules/workflow/)
```

```rust
#[async_trait]
pub trait WorkflowService: Send + Sync {
    async fn start_workflow(&self, req: StartWorkflowRequest) -> Result<WorkflowResponse, WorkflowError>;
    async fn get_status(&self, id: WorkflowId) -> Result<WorkflowStatus, WorkflowError>;
    async fn approve_step(&self, req: ApproveRequest) -> Result<(), WorkflowError>;
}
```

---

# 16. External gRPC (Optional — for SDK/Partner Integrations)

For external integrations (not internal module comms):

```protobuf
// proto/identity/v1/auth_service.proto
package identity.v1;

service AuthService {
    rpc Authenticate(AuthRequest) returns (AuthResponse);
    rpc VerifyToken(VerifyTokenRequest) returns (JWTClaims);
}

// proto/customer/v1/customer_service.proto
package customer.v1;

service CustomerService {
    rpc CreateCustomer(CreateCustomerRequest) returns (Customer);
    rpc GetCustomer(GetCustomerRequest) returns (Customer);
}
```

All gRPC packages include version (`v1`) in the namespace.

---

# 17. Error Standards

All modules use Rust `Result<T, E>`:

```rust
pub enum CommonError {
    Unknown,
    InvalidRequest(String),
    Unauthorized,
    Forbidden,
    NotFound(String),
    Timeout,
    ModelError(String),
    DatabaseError(String),
}
```

---

# 17. NATS JetStream Architecture

NATS is the event backbone.

Used for:

* AI tasks
* Document processing
* Agent lifecycle
* Audit events
* Workflow events

---

# 18. NATS Subject Naming Standard

Format:

```
aeroxe.<domain>.<event>
```

Example:

```
aeroxe.ai.request.created

aeroxe.agent.execution.started

aeroxe.rag.document.processed

aeroxe.vision.analysis.completed

```

---

# 19. Core NATS Subjects

## AI Events

```
aeroxe.ai.request.created

aeroxe.ai.response.generated

aeroxe.ai.failed

```

---

## Agent Events

```
aeroxe.agent.started

aeroxe.agent.completed

aeroxe.agent.failed

```

---

## RAG Events

```
aeroxe.rag.document.uploaded

aeroxe.rag.document.processed

aeroxe.rag.embedding.created

```

---

## Vision Events

```
aeroxe.vision.image.received

aeroxe.vision.analysis.completed

```

---

## Workflow Events

```
aeroxe.workflow.started

aeroxe.workflow.completed

aeroxe.workflow.failed

```

---

## Security Events

```
aeroxe.security.scan.started

aeroxe.security.threat.detected

```

---

# 20. Event Schema Standard

Every event:

```json
{
 "event_id":"uuid",

 "event_type":"AgentCompleted",

 "timestamp":"2026-07-15T12:00:00Z",

 "tenant_id":"uuid",

 "service":"agent-service",

 "version":"1.0",

 "data":{

 }

}

```

---

# 21. Example Event

Agent Completed:

```json
{
 "event_type":"AgentCompleted",

 "service":"agent-orchestrator",

 "data":{

    "execution_id":"12345",

    "agent":"customer-agent",

    "status":"success"

 }

}

```

---

# 22. JetStream Stream Design

## AI Stream

```
AI_EVENTS
```

Subjects:

```
aeroxe.ai.*
```

Retention:

```
7 days
```

---

## Agent Stream

```
AGENT_EVENTS
```

Subjects:

```
aeroxe.agent.*
```

Retention:

```
30 days
```

---

## Audit Stream

```
AUDIT_EVENTS
```

Subjects:

```
aeroxe.audit.*
```

Retention:

```
365 days
```

---

# 23. Request Flow Example

## User asks:

"Why is customer internet slow?"

Flow:

```
User

 |

API Gateway

 |

AI Gateway

 |

Agent Orchestrator

 |

LFM Thinking Model

 |

Plan Created


 |

NATS Event

aeroxe.agent.execution.started


 |

Customer Agent


 |

gRPC


 |

Broadband Service


 |

SQL Agent


 |

Customer DB


 |

RAG Service


 |

Command-R


 |

Final Response


```

---

# 24. Streaming Response Architecture

For Chat UI:

```
User

 |

WebSocket

 |

AI Gateway

 |

gRPC Stream

 |

Ollama


Token Streaming


 |

User Interface

```

---

# 25. Security Requirements

gRPC:

* TLS encryption
* Service authentication
* Metadata validation

NATS:

* TLS
* Account isolation
* Subject permissions

---

# 26. Final Communication Stack

| Layer           | Technology         |
| --------------- | ------------------ |
| Mobile/Web API  | REST               |
| Real-time Chat  | WebSocket          |
| Internal RPC    | gRPC               |
| Contract        | Protocol Buffers   |
| Event Bus       | NATS JetStream     |
| AI Runtime API  | Ollama API         |
| Database Access | Repository Pattern |

---

# Part 3 Completed

The AeroXe Nexus AI communication foundation is now defined:

✅ gRPC Service Contracts
✅ Protobuf Structure
✅ NATS JetStream Event Architecture
✅ Event Naming Standards
✅ Streaming Design
✅ Security Rules
