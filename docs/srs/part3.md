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

               Structured REST / WebSocket / HTTPS (versioned: /api/v1/)

               Standard REST (GET/POST/PUT/PATCH/DELETE)


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
HTTPS Structured REST (/api/v1/*) — Standard REST methods (GET/POST/PUT/PATCH/DELETE)
WebSocket (/ws/v1/*)
```

### Structured REST Standards

| Rule | Description |
|---|---|
| HTTP Method | **Standard REST** (GET, POST, PUT, PATCH, DELETE as appropriate) |
| Path Variables | Resource IDs in request body or URL path |
| Query Strings | Supported for filtering and pagination |
| Request Format | JSON |
| Response Format | JSON |
| Business Status | Every response includes `status` field |
| Pagination | Query parameters or request body: `limit` (default 10) + `offset` |

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

pub struct RequestEnvelope {
    pub operation: String,      // e.g., "GetCustomer", "ListCustomers"
    pub request_id: String,
    pub tenant_id: i64,
    pub user_id: i64,
    pub data: Vec<u8>,          // Serialized operation-specific request
}

pub struct ResponseEnvelope {
    pub status: String,         // e.g., "SUCCESS", "CREATED", "UPDATED"
    pub operation: String,
    pub request_id: String,
    pub data: Vec<u8>,          // Serialized operation-specific response
    pub summary: Option<Summary>,
    pub pagination: Option<Pagination>,
    pub error: Option<ErrorInfo>,
    pub meta: ResponseMeta,
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

# 18. NATS JetStream Architecture

NATS is the event backbone.

Used for:

* AI tasks
* Document processing
* Agent lifecycle
* Audit events
* Workflow events

---

# 19. NATS Subject Naming Standard

Format:

```
aeroxe.v1.<domain>.<event>
```

Example:

```
aeroxe.v1.ai.request.created

aeroxe.v1.agent.execution.started

aeroxe.v1.rag.document.processed

aeroxe.v1.vision.analysis.completed

```

---

# 20. Core NATS Subjects

## AI Events

```
aeroxe.v1.ai.request.created

aeroxe.v1.ai.response.generated

aeroxe.v1.ai.failed

```

---

## Agent Events

```
aeroxe.v1.agent.started

aeroxe.v1.agent.completed

aeroxe.v1.agent.failed

```

---

## RAG Events

```
aeroxe.v1.rag.document.uploaded

aeroxe.v1.rag.document.processed

aeroxe.v1.rag.embedding.created

```

---

## Vision Events

```
aeroxe.v1.vision.image.received

aeroxe.v1.vision.analysis.completed

```

---

## Workflow Events

```
aeroxe.v1.workflow.started

aeroxe.v1.workflow.completed

aeroxe.v1.workflow.failed

```

---

## Security Events

```
aeroxe.v1.security.scan.started

aeroxe.v1.security.threat.detected

```

---

# 21. Event Schema Standard

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

# 22. Example Event

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

# 23. JetStream Stream Design

## AI Stream

```
AI_EVENTS
```

Subjects:

```
aeroxe.v1.ai.*
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
aeroxe.v1.agent.*
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
aeroxe.v1.audit.*
```

Retention:

```
365 days
```

---

# 24. Request Flow Example

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

aeroxe.v1.agent.execution.started


 |

Customer Agent


 |

Trait call


 |

Broadband Service


 |

SQL Agent


 |

Customer DB


 |

RAG Module


 |

Command-R


 |

Final Response


```

---

# 25. Streaming Response Architecture

For Chat UI:

```
User

 |

WebSocket

 |

AI Gateway (trait call to agent module)

 |

Ollama HTTP Streaming API


Token Streaming


 |

User Interface

```

---

# 26. Security Requirements

Structured REST:

* TLS encryption
* JWT authentication
* API key authentication
* Rate limiting
* Request validation

NATS:
* Service authentication
* Metadata validation

NATS:

* TLS
* Account isolation
* Subject permissions

---

# 27. Final Communication Stack

| Layer           | Technology         |
| --------------- | ------------------ |
| Mobile/Web API  | Structured REST (GET/POST/PUT/PATCH/DELETE) |
| Real-time Chat  | WebSocket          |
| Internal RPC    | Rust Trait Interfaces (in-process) |
| Contract        | Protobuf (JSON-serialized) |
| Event Bus       | NATS JetStream     |
| AI Runtime API  | Ollama API         |
| Database Access | Repository Pattern (SeaORM) |

---

# 28. Structured REST API Definitions (NEW Modules)

All operations use standard REST methods (GET/POST/PUT/PATCH/DELETE) with JSON request/response bodies.

## 27.1 Telephony Module API

```protobuf
// Telephony Operations
message InitiateOutboundCallRequest {
  string callee_number = 1;
  string agent_id = 2;
  map<string, string> context = 3;
}

message GetCallRequest {
  string call_id = 1;
}

message ListCallsRequest {
  int32 limit = 1;
  int32 offset = 2;
  string status = 3;
}

message HoldCallRequest {
  string call_id = 1;
}

message TransferCallRequest {
  string call_id = 1;
  string target_agent_id = 2;
  string target_phone = 3;
  string transfer_type = 4;  // blind | attended
  string reason = 5;
}

message EndCallRequest {
  string call_id = 1;
  string reason = 2;
}

message VerifyPinRequest {
  string call_id = 1;
  string pin = 2;
}

message VerifyVoiceRequest {
  string call_id = 1;
  bytes voice_sample = 2;
}

message ListVoicemailsRequest {
  int32 limit = 1;
  int32 offset = 2;
}

message CreateIVRFlowRequest {
  string name = 1;
  repeated IVRNode nodes = 2;
}

message MonitorCallRequest {
  string call_id = 1;
  string action = 2;  // listen | whisper | barge_in
}
```

## 27.2 Conversation Module API

```protobuf
// Conversation Operations
message CreateConversationRequest {
  string channel = 1;  // chat | voice | hybrid
  int64 customer_id = 2;
  string agent_id = 3;
  string initial_message = 4;
}

message GetConversationRequest {
  string conversation_id = 1;
}

message ListConversationsRequest {
  int32 limit = 1;
  int32 offset = 2;
  string state = 3;
}

message AddMessageRequest {
  string conversation_id = 1;
  string role = 2;  // user | assistant | system | tool
  string content = 3;
}

message GetMessagesRequest {
  string conversation_id = 1;
  int32 limit = 2;
  int32 offset = 3;
}

message EndConversationRequest {
  string conversation_id = 1;
  string outcome = 2;  // resolved | escalated | abandoned
  int32 satisfaction_score = 3;
  string summary = 4;
}
```

## 27.3 STT Module API

```protobuf
// STT Operations
message StartSTTSessionRequest {
  string call_id = 1;
  string language = 2;
  string model = 3;
  int32 sample_rate = 4;
  bool enable_punctuation = 5;
  bool enable_speaker_labels = 6;
  bool redact_pii = 7;
}

message SendAudioChunkRequest {
  string session_id = 1;
  bytes data = 2;
  bool is_final = 3;
  uint64 timestamp = 4;
}

message EndSTTSessionRequest {
  string session_id = 1;
}

message TranscribeAudioRequest {
  bytes audio = 1;
  string language = 2;
  string model = 3;
}
```

## 27.4 TTS Module API

```protobuf
// TTS Operations
message SynthesizeSpeechRequest {
  string text = 1;
  string voice_id = 2;
  float speed = 3;
  float pitch = 4;
  float volume = 5;
  string emotion = 6;
}

message SynthesizeSSMLRequest {
  string ssml = 1;
  string voice_id = 2;
}

message ListVoicesRequest {
  int32 limit = 1;
  int32 offset = 2;
  string language = 3;
}

message CloneVoiceRequest {
  string name = 1;
  string source_speaker = 2;
  bytes reference_audio = 3;
  bool consent_recorded = 4;
}
```

## 27.5 Analytics Module API

```protobuf
// Analytics Operations
message GetDashboardRequest {
  string start_date = 1;
  string end_date = 2;
  string granularity = 3;  // minute | hour | day | week | month
}

message GetConversationMetricsRequest {
  string start_date = 1;
  string end_date = 2;
}

message GetCallMetricsRequest {
  string start_date = 1;
  string end_date = 2;
}

message ListAgentMetricsRequest {
  int32 limit = 1;
  int32 offset = 2;
  string start_date = 3;
  string end_date = 4;
}

message GetAgentPerformanceRequest {
  string agent_id = 1;
  string start_date = 2;
  string end_date = 3;
}

message GetCostBreakdownRequest {
  string start_date = 1;
  string end_date = 2;
}
```

## 27.6 Webhook Module API

```protobuf
// Webhook Operations
message CreateWebhookRequest {
  string name = 1;
  string url = 2;
  string secret = 3;
  repeated string event_types = 4;
  repeated string allowed_ips = 5;
}

message GetWebhookRequest {
  string webhook_id = 1;
}

message ListWebhooksRequest {
  int32 limit = 1;
  int32 offset = 2;
}

message UpdateWebhookRequest {
  string webhook_id = 1;
  string name = 2;
  string url = 3;
  repeated string event_types = 4;
  repeated string allowed_ips = 5;
}

message DeleteWebhookRequest {
  string webhook_id = 1;
}

message TestWebhookRequest {
  string webhook_id = 1;
}

message ListDeliveriesRequest {
  string webhook_id = 1;
  int32 limit = 2;
  int32 offset = 3;
}
```

## 27.7 Outbound Module API

```protobuf
// Outbound Operations
message CreateCampaignRequest {
  string name = 1;
  string description = 2;
  string campaign_type = 3;  // voice | chat | email | sms
  string agent_id = 4;
  string script = 5;
  int32 rate_limit = 6;
}

message GetCampaignRequest {
  string campaign_id = 1;
}

message ListCampaignsRequest {
  int32 limit = 1;
  int32 offset = 2;
}

message StartCampaignRequest {
  string campaign_id = 1;
}

message ScheduleCallbackRequest {
  int64 customer_id = 1;
  string phone_number = 2;
  string agent_id = 3;
  string scheduled_at = 4;
  string reason = 5;
}

message AddDNCRequest {
  string phone_number = 1;
  string reason = 2;
}

message ListDNCRequest {
  int32 limit = 1;
  int32 offset = 2;
}
```

## 27.8 Billing Module API

```protobuf
// Billing Operations
message ListPlansRequest {
  // No parameters needed
}

message CreateSubscriptionRequest {
  string plan_id = 1;
  string payment_method_id = 2;
}

message GetSubscriptionRequest {
  string subscription_id = 1;
}

message UpdateSubscriptionRequest {
  string subscription_id = 1;
  string plan_id = 2;
}

message GetUsageRequest {
  string start_date = 1;
  string end_date = 2;
  string service = 3;  // llm | stt | tts | telephony | storage
}

message ListInvoicesRequest {
  int32 limit = 1;
  int32 offset = 2;
  string status = 3;
}

message GetInvoiceRequest {
  string invoice_id = 1;
}

message PayInvoiceRequest {
  string invoice_id = 1;
  string payment_method_id = 2;
}
```

---

# Part 3 Completed

The AeroXe Nexus AI communication foundation is now defined:

✅ Structured REST API (GET/POST/PUT/PATCH/DELETE)
✅ Envelope Request/Response Envelopes
✅ Business Status Codes
✅ Pagination via Request Body
✅ Rust Trait Interfaces (in-process, no gRPC)
✅ NATS JetStream Event Architecture (Outbox Pattern)
✅ Event Naming Standards
✅ Streaming Design
✅ Security Rules
